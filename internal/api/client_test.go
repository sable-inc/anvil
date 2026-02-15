package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sable-inc/anvil/internal/api"
)

func TestClient_Get_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/health" {
			t.Errorf("path = %s, want /health", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer srv.Close()

	client := api.NewClient(srv.URL, "svc_test")

	var resp struct {
		Status string `json:"status"`
	}
	if err := client.Get(context.Background(), "/health", &resp); err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if resp.Status != "ok" {
		t.Errorf("status = %q, want %q", resp.Status, "ok")
	}
}

func TestClient_Post_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}

		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "test-agent" {
			t.Errorf("body.name = %q, want %q", body["name"], "test-agent")
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{"id": 1, "name": "test-agent"})
	}))
	defer srv.Close()

	client := api.NewClient(srv.URL, "svc_test")

	var resp struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	err := client.Post(context.Background(), "/agents", map[string]string{"name": "test-agent"}, &resp)
	if err != nil {
		t.Fatalf("Post() error: %v", err)
	}
	if resp.ID != 1 {
		t.Errorf("id = %d, want 1", resp.ID)
	}
}

func TestClient_ErrorResponse_Typed(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
		check  func(error) bool
		label  string
	}{
		{
			"401",
			401,
			`{"error":"Unauthorized","hint":"re-login"}`,
			func(e error) bool { _, ok := api.As[*api.UnauthorizedError](e); return ok },
			"UnauthorizedError",
		},
		{
			"404",
			404,
			`{"error":"Agent not found"}`,
			func(e error) bool { _, ok := api.As[*api.NotFoundError](e); return ok },
			"NotFoundError",
		},
		{
			"422",
			422,
			`{"error":"Validation failed","hint":"check name"}`,
			func(e error) bool { _, ok := api.As[*api.ValidationError](e); return ok },
			"ValidationError",
		},
		{
			"500",
			500,
			`{"error":"Internal Server Error"}`,
			func(e error) bool {
				_, nf := api.As[*api.NotFoundError](e)
				_, ua := api.As[*api.UnauthorizedError](e)
				return !nf && !ua // Should not match specific types
			},
			"generic ResponseError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.status)
				w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			client := api.NewClient(srv.URL, "svc_test")
			err := client.Get(context.Background(), "/test", nil)

			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.check(err) {
				t.Errorf("error type check failed for %s: %v", tt.label, err)
			}
		})
	}
}

func TestClient_ErrorResponse_WithHint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(401)
		w.Write([]byte(`{"error":"Unauthorized","hint":"Token signature invalid - please log out and log back in"}`))
	}))
	defer srv.Close()

	client := api.NewClient(srv.URL, "svc_test")
	err := client.Get(context.Background(), "/test", nil)

	if err == nil {
		t.Fatal("expected error")
	}

	ua, ok := api.As[*api.UnauthorizedError](err)
	if !ok {
		t.Fatalf("expected UnauthorizedError, got %T", err)
	}
	if ua.Hint != "Token signature invalid - please log out and log back in" {
		t.Errorf("Hint = %q", ua.Hint)
	}
}

func TestClient_NilResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	client := api.NewClient(srv.URL, "svc_test")

	// nil target should not error on 204 No Content.
	if err := client.Delete(context.Background(), "/test", nil); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
}

func TestClient_TrailingSlash(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/agents" {
			t.Errorf("path = %q, want /agents", r.URL.Path)
		}
		w.WriteHeader(200)
		w.Write([]byte("[]"))
	}))
	defer srv.Close()

	// Base URL with trailing slash should still produce clean paths.
	client := api.NewClient(srv.URL+"/", "svc_test")
	var agents []any
	if err := client.Get(context.Background(), "/agents", &agents); err != nil {
		t.Fatalf("Get() error: %v", err)
	}
}
