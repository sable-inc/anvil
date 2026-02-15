package output_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/sable-inc/anvil/internal/output"
)

func TestTable_Render(t *testing.T) {
	tbl := output.NewTable("ID", "Name", "Status")
	tbl.AddRow("1", "Alpha", "active")
	tbl.AddRow("2", "Beta", "inactive")

	var buf bytes.Buffer
	if err := tbl.Render(&buf); err != nil {
		t.Fatalf("Render error: %v", err)
	}

	out := buf.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")

	if len(lines) != 3 {
		t.Fatalf("expected 3 lines (header + 2 rows), got %d:\n%s", len(lines), out)
	}

	// Header should be uppercase.
	if !strings.Contains(lines[0], "ID") || !strings.Contains(lines[0], "NAME") || !strings.Contains(lines[0], "STATUS") {
		t.Errorf("header = %q, want ID NAME STATUS", lines[0])
	}

	// Rows should contain data.
	if !strings.Contains(lines[1], "Alpha") {
		t.Errorf("row 1 = %q, want Alpha", lines[1])
	}
	if !strings.Contains(lines[2], "Beta") {
		t.Errorf("row 2 = %q, want Beta", lines[2])
	}
}

func TestTable_Empty(t *testing.T) {
	tbl := output.NewTable("ID", "Name")

	var buf bytes.Buffer
	if err := tbl.Render(&buf); err != nil {
		t.Fatalf("Render error: %v", err)
	}

	out := buf.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")

	// Should still have the header row.
	if len(lines) != 1 {
		t.Fatalf("expected 1 line (header only), got %d", len(lines))
	}
}

func TestWrite_JSON(t *testing.T) {
	data := map[string]string{"key": "value"}
	tbl := output.NewTable("Key")
	tbl.AddRow("value")

	var buf bytes.Buffer
	if err := output.Write(&buf, "json", data, tbl); err != nil {
		t.Fatalf("Write error: %v", err)
	}

	var parsed map[string]string
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("JSON parse error: %v", err)
	}
	if parsed["key"] != "value" {
		t.Errorf("key = %q, want value", parsed["key"])
	}
}

func TestWrite_Table(t *testing.T) {
	data := map[string]string{"key": "value"}
	tbl := output.NewTable("Key")
	tbl.AddRow("value")

	var buf bytes.Buffer
	if err := output.Write(&buf, "table", data, tbl); err != nil {
		t.Fatalf("Write error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "KEY") {
		t.Errorf("table should contain header KEY, got: %q", out)
	}
	if !strings.Contains(out, "value") {
		t.Errorf("table should contain value, got: %q", out)
	}
}
