package main

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestWriteJSONFindingsUsesVersionedStableEnvelope(t *testing.T) {
	var output bytes.Buffer
	want := []jsonFinding{{
		SourceFile: "settings.json",
		Event:      "PreToolUse",
		Severity:   "WARN",
		Code:       "matcher_typo",
		Args:       []any{"Bahs", "Bash"},
	}}

	if err := writeJSONFindings(&output, want); err != nil {
		t.Fatalf("writeJSONFindings() error = %v", err)
	}

	var got struct {
		Version  int           `json:"version"`
		Findings []jsonFinding `json:"findings"`
	}
	if err := json.Unmarshal(output.Bytes(), &got); err != nil {
		t.Fatalf("JSON output is invalid: %v", err)
	}
	if got.Version != 1 {
		t.Errorf("version = %d, want 1", got.Version)
	}
	if len(got.Findings) != 1 || got.Findings[0].Code != "matcher_typo" {
		t.Errorf("findings = %#v, want one matcher_typo finding", got.Findings)
	}
}

func TestWriteJSONFindingsEmitsEmptyArray(t *testing.T) {
	var output bytes.Buffer
	if err := writeJSONFindings(&output, []jsonFinding{}); err != nil {
		t.Fatalf("writeJSONFindings() error = %v", err)
	}
	if got := output.String(); got != `{"version":1,"findings":[]}`+"\n" {
		t.Errorf("empty output = %q, want versioned empty array", got)
	}
}
