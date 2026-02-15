package configascode

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// FromYAMLFile reads and parses a YAML config file from disk.
func FromYAMLFile(path string) (*ConfigFile, error) {
	data, err := os.ReadFile(path) //nolint:gosec // user-specified config file
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}
	return FromYAML(data)
}

// FromYAML parses a ConfigFile from YAML bytes.
func FromYAML(data []byte) (*ConfigFile, error) {
	var cfg ConfigFile
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}
	return &cfg, nil
}

// ToYAML serializes a ConfigFile to YAML bytes.
func ToYAML(cfg *ConfigFile) ([]byte, error) {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("marshaling YAML: %w", err)
	}
	return data, nil
}

// WriteYAML writes a ConfigFile as YAML to the given writer.
func WriteYAML(w io.Writer, cfg *ConfigFile) error {
	data, err := ToYAML(cfg)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

// FromJSON parses an AgentConfig from JSON bytes (API response format).
func FromJSON(data []byte) (*AgentConfig, error) {
	var cfg AgentConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}
	return &cfg, nil
}

// ToJSON serializes an AgentConfig to JSON bytes.
func ToJSON(cfg *AgentConfig) ([]byte, error) {
	data, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("marshaling JSON: %w", err)
	}
	return data, nil
}
