package configascode_test

import (
	"bytes"
	"testing"

	"github.com/sable-inc/anvil/internal/configascode"
)

func TestDiff_Identical(t *testing.T) {
	cfg := validConfig()
	result, err := configascode.Diff(&cfg, &cfg)
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}
	if result.HasChanges() {
		t.Error("expected no changes for identical configs")
	}
}

func TestDiff_ChangedField(t *testing.T) {
	local := validConfig()
	remote := validConfig()
	remote.Name = "Different Name"

	result, err := configascode.Diff(&local, &remote)
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}
	if !result.HasChanges() {
		t.Fatal("expected changes")
	}
	if len(result.Changed) == 0 {
		t.Fatal("expected at least one changed field")
	}

	found := false
	for _, d := range result.Changed {
		if d.Path == "name" {
			found = true
			if d.Local != "Test Agent" {
				t.Errorf("local name = %v, want Test Agent", d.Local)
			}
			if d.Remote != "Different Name" {
				t.Errorf("remote name = %v, want Different Name", d.Remote)
			}
		}
	}
	if !found {
		t.Error("expected 'name' in changed fields")
	}
}

func TestDiff_NestedChange(t *testing.T) {
	local := validConfig()
	remote := validConfig()
	remote.LLM.Temperature = 1.5

	result, err := configascode.Diff(&local, &remote)
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}
	if !result.HasChanges() {
		t.Fatal("expected changes")
	}

	found := false
	for _, d := range result.Changed {
		if d.Path == "llm.temperature" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'llm.temperature' in changed fields")
	}
}

func TestWriteDiff_NoChanges(t *testing.T) {
	result := &configascode.DiffResult{}
	var buf bytes.Buffer
	if err := configascode.WriteDiff(&buf, result); err != nil {
		t.Fatalf("WriteDiff: %v", err)
	}
	if buf.String() != "No differences found.\n" {
		t.Errorf("unexpected output: %q", buf.String())
	}
}

func TestSummaryLine_NoChanges(t *testing.T) {
	result := &configascode.DiffResult{}
	if s := configascode.SummaryLine(result); s != "configs are identical" {
		t.Errorf("summary = %q, want 'configs are identical'", s)
	}
}

func TestSummaryLine_WithChanges(t *testing.T) {
	result := &configascode.DiffResult{
		Changed: []configascode.FieldDiff{{Path: "name"}},
		Added:   []configascode.FieldDiff{{Path: "new_field"}, {Path: "another"}},
	}
	s := configascode.SummaryLine(result)
	if s != "1 changed, 2 added" {
		t.Errorf("summary = %q, want '1 changed, 2 added'", s)
	}
}
