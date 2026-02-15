package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sable-inc/anvil/internal/api"
)

func TestAuthTransport_InjectsHeader(t *testing.T) {
	var gotHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("Authorization")
	}))
	defer srv.Close()

	client := &http.Client{
		Transport: &api.AuthTransport{Token: "svc_my_token"},
	}

	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	want := "Bearer svc_my_token"
	if gotHeader != want {
		t.Errorf("Authorization = %q, want %q", gotHeader, want)
	}
}

func TestAuthTransport_EmptyToken(t *testing.T) {
	var gotHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("Authorization")
	}))
	defer srv.Close()

	client := &http.Client{
		Transport: &api.AuthTransport{Token: ""},
	}

	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if gotHeader != "" {
		t.Errorf("Authorization should be empty with no token, got %q", gotHeader)
	}
}

func TestAuthTransport_PreservesExistingHeaders(t *testing.T) {
	var gotAccept string
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotAccept = r.Header.Get("Accept")
	}))
	defer srv.Close()

	client := &http.Client{
		Transport: &api.AuthTransport{Token: "svc_test"},
	}

	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if gotAccept != "application/json" {
		t.Errorf("Accept = %q, want %q", gotAccept, "application/json")
	}
}
