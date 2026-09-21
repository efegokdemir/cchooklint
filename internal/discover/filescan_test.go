package discover

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFind(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	home := filepath.Join(root, "home")
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	projectDir := filepath.Join(root, ".claude")
	homeDir := filepath.Join(home, ".claude")
	for _, dir := range []string{projectDir, homeDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("MkdirAll(%q): %v", dir, err)
		}
	}

	want := []string{
		filepath.Join(".claude", "settings.json"),
		filepath.Join(home, ".claude", "settings.json"),
	}
	if err := os.WriteFile(want[0], nil, 0o600); err != nil {
		t.Fatalf("WriteFile(%q): %v", want[0], err)
	}
	if err := os.WriteFile(want[1], nil, 0o600); err != nil {
		t.Fatalf("WriteFile(%q): %v", want[1], err)
	}

	got, err := Find()
	if err != nil {
		t.Fatalf("Find() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Find() = %v, want %v", got, want)
	}
}

func TestFindExcludesMissingFiles(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", os.Getenv("HOME"))

	got, err := Find()
	if err != nil {
		t.Fatalf("Find() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("Find() = %v, want no settings files", got)
	}
}

func TestHomeSettingsPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	got, err := homeSettingsPath()
	if err != nil {
		t.Fatalf("homeSettingsPath() error = %v", err)
	}
	want := filepath.Join(home, ".claude", "settings.json")
	if got != want {
		t.Errorf("homeSettingsPath() = %q, want %q", got, want)
	}
}
