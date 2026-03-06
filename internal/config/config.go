// Package config handles CLI configuration loading from ~/.config/anvil/config.yaml.
package config

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the Anvil CLI configuration file.
type Config struct {
	DefaultOrg    string               `yaml:"default_org"`
	APIURL        string               `yaml:"api_url"`
	Format        string               `yaml:"format"`
	Orgs          map[string]OrgConfig `yaml:"orgs"`
	HyperDXAPIKey string               `yaml:"hyperdx_api_key,omitempty"`
	HyperDXAPIURL string               `yaml:"hyperdx_api_url,omitempty"`
}

// OrgConfig stores per-organization overrides.
type OrgConfig struct {
	Name   string `yaml:"name"`
	APIURL string `yaml:"api_url"`
}

// Dir returns the XDG-compliant config directory (~/.config/anvil).
func Dir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "anvil")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".config", "anvil")
	}
	return filepath.Join(home, ".config", "anvil")
}

// Path returns the full path to config.yaml.
func Path() string {
	return filepath.Join(Dir(), "config.yaml")
}

// Load reads and parses the config file.
// Returns a zero-value Config if the file doesn't exist.
func Load() (*Config, error) {
	data, err := os.ReadFile(Path())
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return &Config{}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Save writes the config to disk, creating directories as needed.
func Save(cfg *Config) error {
	if err := os.MkdirAll(Dir(), 0o750); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(Path(), data, 0o600)
}
