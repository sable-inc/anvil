package auth_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/sable-inc/anvil/internal/auth"
)

func TestSaveAndLoad(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	if err := auth.Save("svc_test_token_123"); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Verify the file was created with correct permissions.
	credPath := filepath.Join(tmp, "anvil", "credentials.json")
	info, err := os.Stat(credPath)
	if err != nil {
		t.Fatalf("credentials file not found: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("file permissions = %o, want 0600", perm)
	}

	// Verify contents.
	data, err := os.ReadFile(credPath) //nolint:gosec // test reads from t.TempDir()
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("JSON parse error: %v", err)
	}
	if raw["token"] != "svc_test_token_123" {
		t.Errorf("token = %q, want %q", raw["token"], "svc_test_token_123")
	}
	if raw["created_at"] == "" {
		t.Error("created_at should not be empty")
	}

	// Load it back.
	creds, err := auth.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if creds == nil {
		t.Fatal("Load() returned nil")
	}
	if creds.Token != "svc_test_token_123" {
		t.Errorf("Token = %q, want %q", creds.Token, "svc_test_token_123")
	}
}

func TestLoad_NoFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	creds, err := auth.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if creds != nil {
		t.Errorf("Load() should return nil when no file exists, got %+v", creds)
	}
}

func TestClear(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	if err := auth.Save("svc_to_delete"); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	if err := auth.Clear(); err != nil {
		t.Fatalf("Clear() error: %v", err)
	}

	creds, err := auth.Load()
	if err != nil {
		t.Fatalf("Load() after Clear() error: %v", err)
	}
	if creds != nil {
		t.Error("Load() should return nil after Clear()")
	}
}

func TestClear_NoFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	// Clear on non-existent file should not error.
	if err := auth.Clear(); err != nil {
		t.Fatalf("Clear() on non-existent file should not error, got: %v", err)
	}
}
