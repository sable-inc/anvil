// Package auth manages CLI authentication credentials.
package auth

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// Credentials holds a stored authentication token.
type Credentials struct {
	Token     string `json:"token"`
	CreatedAt string `json:"created_at"`
}

// Dir returns the XDG-compliant config directory for anvil credentials.
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

// Path returns the full path to the credentials file.
func Path() string {
	return filepath.Join(Dir(), "credentials.json")
}

// Load reads stored credentials from disk.
// Returns nil, nil if the file doesn't exist (not an error — just not logged in).
func Load() (*Credentials, error) {
	data, err := os.ReadFile(Path())
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, err
	}
	return &creds, nil
}

// Save writes credentials to disk, creating directories as needed.
func Save(token string) error {
	if err := os.MkdirAll(Dir(), 0o750); err != nil {
		return err
	}

	creds := Credentials{
		Token:     token,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(Path(), data, 0o600)
}

// Clear removes stored credentials.
// Returns nil if the file doesn't exist.
func Clear() error {
	err := os.Remove(Path())
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return err
}
