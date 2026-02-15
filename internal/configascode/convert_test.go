package configascode_test

import (
	"bytes"
	"testing"

	"github.com/sable-inc/anvil/internal/configascode"
)

func TestYAMLRoundTrip(t *testing.T) {
	cfg := validConfig()
	cfgFile := &configascode.ConfigFile{
		Agent:  "test-agent",
		OrgID:  42,
		Config: cfg,
	}

	data, err := configascode.ToYAML(cfgFile)
	if err != nil {
		t.Fatalf("ToYAML: %v", err)
	}

	parsed, err := configascode.FromYAML(data)
	if err != nil {
		t.Fatalf("FromYAML: %v", err)
	}

	if parsed.Agent != "test-agent" {
		t.Errorf("agent = %q, want test-agent", parsed.Agent)
	}
	if parsed.OrgID != 42 {
		t.Errorf("org_id = %d, want 42", parsed.OrgID)
	}
	if parsed.Config.Name != "Test Agent" {
		t.Errorf("config.name = %q, want 'Test Agent'", parsed.Config.Name)
	}
	if parsed.Config.TTS.Provider != "elevenlabs" {
		t.Errorf("config.tts.provider = %q, want elevenlabs", parsed.Config.TTS.Provider)
	}
}

func TestJSONRoundTrip(t *testing.T) {
	cfg := validConfig()
	data, err := configascode.ToJSON(&cfg)
	if err != nil {
		t.Fatalf("ToJSON: %v", err)
	}

	parsed, err := configascode.FromJSON(data)
	if err != nil {
		t.Fatalf("FromJSON: %v", err)
	}

	if parsed.Name != cfg.Name {
		t.Errorf("name = %q, want %q", parsed.Name, cfg.Name)
	}
	if parsed.STT.Provider != cfg.STT.Provider {
		t.Errorf("stt.provider = %q, want %q", parsed.STT.Provider, cfg.STT.Provider)
	}
}

func TestWriteYAML(t *testing.T) {
	cfg := validConfig()
	cfgFile := &configascode.ConfigFile{Config: cfg}

	var buf bytes.Buffer
	if err := configascode.WriteYAML(&buf, cfgFile); err != nil {
		t.Fatalf("WriteYAML: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("expected non-empty YAML output")
	}
}

func TestFromYAMLFile_NotFound(t *testing.T) {
	_, err := configascode.FromYAMLFile("/nonexistent/path.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestFromYAML_InvalidYAML(t *testing.T) {
	_, err := configascode.FromYAML([]byte("{{invalid"))
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}
