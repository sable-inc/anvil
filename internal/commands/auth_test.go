package commands_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sable-inc/anvil/internal/auth"
	"github.com/sable-inc/anvil/internal/commands"
)

func TestAuthLogin(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	root := commands.NewRoot()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"auth", "login", "--token", "svc_test_token_abc"})

	if err := root.Execute(); err != nil {
		t.Fatalf("auth login error: %v", err)
	}

	if !strings.Contains(out.String(), "Logged in successfully") {
		t.Errorf("output = %q, want 'Logged in successfully'", out.String())
	}

	// Verify credentials were saved.
	creds, err := auth.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if creds == nil {
		t.Fatal("credentials should exist after login")
	}
	if creds.Token != "svc_test_token_abc" {
		t.Errorf("Token = %q, want %q", creds.Token, "svc_test_token_abc")
	}
}

func TestAuthLogout(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	// Save credentials first.
	if err := auth.Save("svc_to_remove"); err != nil {
		t.Fatal(err)
	}

	root := commands.NewRoot()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"auth", "logout"})

	if err := root.Execute(); err != nil {
		t.Fatalf("auth logout error: %v", err)
	}

	if !strings.Contains(out.String(), "Logged out") {
		t.Errorf("output = %q, want 'Logged out'", out.String())
	}

	creds, _ := auth.Load()
	if creds != nil {
		t.Error("credentials should not exist after logout")
	}
}

func TestAuthWhoami_NotLoggedIn(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	root := commands.NewRoot()
	root.SetArgs([]string{"auth", "whoami"})

	err := root.Execute()
	if err == nil {
		t.Fatal("whoami should error when not logged in")
	}
}

func TestAuthWhoami_LoggedIn(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	if err := auth.Save("svc_test_token_xyz"); err != nil {
		t.Fatal(err)
	}

	root := commands.NewRoot()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"auth", "whoami", "--format", "json"})

	if err := root.Execute(); err != nil {
		t.Fatalf("auth whoami error: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("JSON parse error: %v (output: %q)", err, out.String())
	}

	// Token should be masked.
	if !strings.Contains(result["token"], "...") {
		t.Errorf("token should be masked, got %q", result["token"])
	}
	if result["path"] == "" {
		t.Error("path should not be empty")
	}
}

func TestAuthStatus_Valid(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	// Mock server that accepts auth.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/health":
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		case "/agents":
			if authHdr := r.Header.Get("Authorization"); authHdr != "Bearer svc_valid_token" {
				w.WriteHeader(401)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
				return
			}
			_ = json.NewEncoder(w).Encode([]any{})
		default:
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	if err := auth.Save("svc_valid_token"); err != nil {
		t.Fatal(err)
	}

	root := commands.NewRoot()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"auth", "status", "--api-url", srv.URL})

	if err := root.Execute(); err != nil {
		t.Fatalf("auth status error: %v", err)
	}

	if !strings.Contains(out.String(), "Authenticated") {
		t.Errorf("output = %q, want 'Authenticated'", out.String())
	}
}

func TestAuthStatus_InvalidToken(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/health":
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		case "/agents":
			w.WriteHeader(401)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		}
	}))
	defer srv.Close()

	if err := auth.Save("svc_bad_token"); err != nil {
		t.Fatal(err)
	}

	root := commands.NewRoot()
	root.SetArgs([]string{"auth", "status", "--api-url", srv.URL})

	err := root.Execute()
	if err == nil {
		t.Fatal("auth status should error with invalid token")
	}
	if !strings.Contains(err.Error(), "invalid or expired") {
		t.Errorf("error = %q, want message about invalid/expired token", err.Error())
	}
}

func TestAuthStatus_NotLoggedIn(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	root := commands.NewRoot()
	root.SetArgs([]string{"auth", "status", "--api-url", "http://localhost:9999"})

	err := root.Execute()
	if err == nil {
		t.Fatal("auth status should error when not logged in")
	}
	if !strings.Contains(err.Error(), "not authenticated") {
		t.Errorf("error = %q, want 'not authenticated'", err.Error())
	}
}
