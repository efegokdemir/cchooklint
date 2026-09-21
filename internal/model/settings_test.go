package model

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestLoadParsesSettingsAndReportsMissingFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	contents := `{"hooks":{"PreToolUse":[{"matcher":"Bash|PowerShell","hooks":[{"type":"command","command":"./guard.sh"}]}]}}`
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	want := Settings{Hooks: map[string][]HookMatcherGroup{
		"PreToolUse": {{Matcher: "Bash|PowerShell", Hooks: []HookCommand{{Type: "command", Command: "./guard.sh"}}}},
	}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Load() = %#v, want %#v", got, want)
	}

	if _, err := Load(filepath.Join(dir, "missing.json")); err == nil {
		t.Error("Load(missing file) returned nil error")
	}
}

func TestFlattenExpandsMatcherGroupsAndCommands(t *testing.T) {
	t.Parallel()

	settings := Settings{Hooks: map[string][]HookMatcherGroup{
		"PreToolUse": {
			{Matcher: "Bash", Hooks: []HookCommand{{Type: "command", Command: "./first.sh"}, {Command: "./second.sh"}}},
			{Matcher: "PowerShell", Hooks: []HookCommand{{Command: "./third.ps1"}}},
		},
	}}
	got := Flatten("project/settings.json", settings)
	sort.Slice(got, func(i, j int) bool { return got[i].Command < got[j].Command })
	want := []HookEntry{
		{SourceFile: "project/settings.json", Event: "PreToolUse", Matcher: "Bash", Command: "./first.sh"},
		{SourceFile: "project/settings.json", Event: "PreToolUse", Matcher: "Bash", Command: "./second.sh"},
		{SourceFile: "project/settings.json", Event: "PreToolUse", Matcher: "PowerShell", Command: "./third.ps1"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Flatten() = %#v, want %#v", got, want)
	}
}

func TestSplitMatcher(t *testing.T) {
	t.Parallel()

	tests := map[string][]string{
		"Bash|PowerShell,WebFetch": {"Bash", "PowerShell", "WebFetch"},
		"Bash":                     {"Bash"},
		"":                         {},
		"Bash||PowerShell,":        {"Bash", "PowerShell"},
	}
	for input, want := range tests {
		if got := SplitMatcher(input); !reflect.DeepEqual(got, want) {
			t.Errorf("SplitMatcher(%q) = %#v, want %#v", input, got, want)
		}
	}
}

func TestIsRegexMatcher(t *testing.T) {
	t.Parallel()

	tests := map[string]bool{
		"Bash|PowerShell":     false,
		"Bash.*":              true,
		"^(Bash|PowerShell)$": true,
		"WebFetch":            false,
	}
	for input, want := range tests {
		if got := IsRegexMatcher(input); got != want {
			t.Errorf("IsRegexMatcher(%q) = %t, want %t", input, got, want)
		}
	}
}

func TestEditDistance(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		a, b string
		want int
	}{
		"identical":    {a: "Bash", b: "Bash", want: 0},
		"substitution": {a: "Bash", b: "Bahs", want: 2},
		"insertion":    {a: "Bash", b: "Bashh", want: 1},
		"deletion":     {a: "Bash", b: "Bas", want: 1},
		"different":    {a: "abc", b: "xyz", want: 3},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := EditDistance(tt.a, tt.b); got != tt.want {
				t.Errorf("EditDistance(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
