package output_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/sable-inc/anvil/internal/output"
)

func TestJSONFormatter(t *testing.T) {
	var buf bytes.Buffer
	f := output.New("json")

	data := map[string]string{"name": "test"}
	if err := f.Format(&buf, data); err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if result["name"] != "test" {
		t.Errorf("name = %q, want %q", result["name"], "test")
	}
}

func TestYAMLFormatter(t *testing.T) {
	var buf bytes.Buffer
	f := output.New("yaml")

	data := map[string]string{"name": "test"}
	if err := f.Format(&buf, data); err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	if !strings.Contains(buf.String(), "name: test") {
		t.Errorf("output = %q, want to contain %q", buf.String(), "name: test")
	}
}

func TestTableFormatter_FallsBackToJSON(t *testing.T) {
	var buf bytes.Buffer
	f := output.New("table")

	data := map[string]string{"name": "test"}
	if err := f.Format(&buf, data); err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	if !strings.Contains(buf.String(), "name") {
		t.Errorf("output should contain data, got: %q", buf.String())
	}
}

func TestNew_UnknownFormat_DefaultsToJSON(t *testing.T) {
	var buf bytes.Buffer
	f := output.New("unknown")

	data := map[string]string{"key": "value"}
	if err := f.Format(&buf, data); err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("unknown format should default to JSON, got: %q", buf.String())
	}
}
