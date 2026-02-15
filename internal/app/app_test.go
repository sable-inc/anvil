package app_test

import (
	"bytes"
	"testing"

	"github.com/sable-inc/anvil/internal/app"
)

func TestNew_Defaults(t *testing.T) {
	a := app.New()

	if a.Format != "table" {
		t.Errorf("default Format = %q, want %q", a.Format, "table")
	}
	if a.Out == nil {
		t.Error("Out should not be nil")
	}
	if a.ErrOut == nil {
		t.Error("ErrOut should not be nil")
	}
	if a.Verbose {
		t.Error("Verbose should default to false")
	}
}

func TestNew_WithOptions(t *testing.T) {
	var buf bytes.Buffer
	a := app.New(
		app.WithOutput(&buf),
		app.WithFormat("json"),
		app.WithAPIURL("https://api.example.com"),
		app.WithOrgID("org-123"),
		app.WithVerbose(true),
		app.WithNoColor(true),
	)

	if a.Format != "json" {
		t.Errorf("Format = %q, want %q", a.Format, "json")
	}
	if a.APIURL != "https://api.example.com" {
		t.Errorf("APIURL = %q, want %q", a.APIURL, "https://api.example.com")
	}
	if a.OrgID != "org-123" {
		t.Errorf("OrgID = %q, want %q", a.OrgID, "org-123")
	}
	if !a.Verbose {
		t.Error("Verbose should be true")
	}
	if !a.NoColor {
		t.Error("NoColor should be true")
	}
}
