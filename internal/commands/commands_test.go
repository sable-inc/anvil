package commands_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sable-inc/anvil/internal/auth"
	"github.com/sable-inc/anvil/internal/commands"
)

// testEnv sets up a temp XDG dir, stores a token, and returns a teardown-free test harness.
func testEnv(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	if err := auth.Save("svc_test_token"); err != nil {
		t.Fatal(err)
	}
}

// runCmd executes a command and returns stdout, stderr, and any error.
func runCmd(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	root := commands.NewRoot()
	var out, errBuf bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errBuf)
	root.SetArgs(args)
	err = root.Execute()
	return out.String(), errBuf.String(), err
}

// jsonHandler returns an http.HandlerFunc that responds with JSON.
func jsonHandler(routes map[string]any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		data, ok := routes[r.Method+" "+r.URL.Path]
		if !ok {
			w.WriteHeader(404)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Not Found"})
			return
		}
		_ = json.NewEncoder(w).Encode(data)
	}
}

// --- Health ---

func TestHealth(t *testing.T) {
	srv := httptest.NewServer(jsonHandler(map[string]any{
		"GET /health": map[string]string{"status": "ok"},
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "health", "--api-url", srv.URL)
	if err != nil {
		t.Fatalf("health error: %v", err)
	}
	if out == "" {
		t.Error("expected output")
	}
}

func TestHealth_WithDB(t *testing.T) {
	srv := httptest.NewServer(jsonHandler(map[string]any{
		"GET /health":    map[string]string{"status": "ok"},
		"GET /db/health": map[string]string{"status": "ok"},
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "health", "--db", "--api-url", srv.URL)
	if err != nil {
		t.Fatalf("health --db error: %v", err)
	}
	if out == "" {
		t.Error("expected output")
	}
}

func TestHealth_JSON(t *testing.T) {
	srv := httptest.NewServer(jsonHandler(map[string]any{
		"GET /health": map[string]string{"status": "ok"},
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "health", "--api-url", srv.URL, "--format", "json")
	if err != nil {
		t.Fatalf("health --format json error: %v", err)
	}

	var data map[string]string
	if err := json.Unmarshal([]byte(out), &data); err != nil {
		t.Fatalf("JSON parse error: %v", err)
	}
	if data["api"] != "ok" {
		t.Errorf("api = %q, want ok", data["api"])
	}
}

// --- Agent ---

func TestAgentList(t *testing.T) {
	agents := []map[string]any{
		{"id": 1, "publicId": "agt_abc", "name": "Test Agent", "slug": "test-agent", "status": "active"},
		{"id": 2, "publicId": "agt_def", "name": "Another", "slug": "another", "status": "inactive"},
	}
	srv := httptest.NewServer(jsonHandler(map[string]any{
		"GET /agents": map[string]any{"agents": agents},
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "agent", "list", "--api-url", srv.URL)
	if err != nil {
		t.Fatalf("agent list error: %v", err)
	}
	if out == "" {
		t.Error("expected output")
	}
}

func TestAgentList_JSON(t *testing.T) {
	agents := []map[string]any{
		{"id": 1, "publicId": "agt_abc", "name": "Test", "slug": "test", "status": "active"},
	}
	srv := httptest.NewServer(jsonHandler(map[string]any{
		"GET /agents": map[string]any{"agents": agents},
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "agent", "list", "--api-url", srv.URL, "--format", "json")
	if err != nil {
		t.Fatalf("agent list --format json error: %v", err)
	}

	var data []map[string]any
	if err := json.Unmarshal([]byte(out), &data); err != nil {
		t.Fatalf("JSON parse: %v", err)
	}
	if len(data) != 1 {
		t.Errorf("expected 1 agent, got %d", len(data))
	}
}

func TestAgentGet(t *testing.T) {
	agent := map[string]any{
		"id": 1, "publicId": "agt_abc", "orgId": 1,
		"name": "Test", "slug": "test", "status": "active",
		"createdAt": "2025-01-01T00:00:00Z", "updatedAt": "2025-01-01T00:00:00Z",
	}
	srv := httptest.NewServer(jsonHandler(map[string]any{
		"GET /agents/test": map[string]any{"agent": agent},
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "agent", "get", "test", "--api-url", srv.URL, "--format", "json")
	if err != nil {
		t.Fatalf("agent get error: %v", err)
	}

	var data map[string]any
	if err := json.Unmarshal([]byte(out), &data); err != nil {
		t.Fatalf("JSON parse: %v", err)
	}
	if data["name"] != "Test" {
		t.Errorf("name = %v, want Test", data["name"])
	}
}

func TestAgentCreate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "POST" && r.URL.Path == "/agents" {
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			resp := map[string]any{"agent": map[string]any{
				"id": 99, "publicId": "agt_new", "orgId": 1,
				"name": body["name"], "slug": "new-agent", "status": "active",
				"createdAt": "2025-01-01T00:00:00Z", "updatedAt": "2025-01-01T00:00:00Z",
			}}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		w.WriteHeader(404)
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "agent", "create", "--name", "New Agent", "--api-url", srv.URL, "--format", "json")
	if err != nil {
		t.Fatalf("agent create error: %v", err)
	}

	var data map[string]any
	if err := json.Unmarshal([]byte(out), &data); err != nil {
		t.Fatalf("JSON parse: %v", err)
	}
	if data["name"] != "New Agent" {
		t.Errorf("name = %v, want 'New Agent'", data["name"])
	}
}

// --- Journey ---

func TestJourneyList(t *testing.T) {
	journeys := []map[string]any{
		{"id": 1, "publicId": "jny_abc", "name": "Onboarding", "slug": "onboarding", "version": 3},
	}
	srv := httptest.NewServer(jsonHandler(map[string]any{
		"GET /journeys": map[string]any{"journeys": journeys},
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "journey", "list", "--api-url", srv.URL)
	if err != nil {
		t.Fatalf("journey list error: %v", err)
	}
	if out == "" {
		t.Error("expected output")
	}
}

func TestJourneyGet(t *testing.T) {
	journey := map[string]any{
		"id": 1, "publicId": "jny_abc", "orgId": 1,
		"name": "Onboarding", "slug": "onboarding", "description": "Welcome flow",
		"version": 3, "moments": []any{}, "transitions": []any{},
		"createdAt": "2025-01-01T00:00:00Z", "updatedAt": "2025-01-01T00:00:00Z",
	}
	srv := httptest.NewServer(jsonHandler(map[string]any{
		"GET /journeys/1": map[string]any{"journey": journey},
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "journey", "get", "1", "--api-url", srv.URL, "--format", "json")
	if err != nil {
		t.Fatalf("journey get error: %v", err)
	}

	var data map[string]any
	if err := json.Unmarshal([]byte(out), &data); err != nil {
		t.Fatalf("JSON parse: %v", err)
	}
	if data["name"] != "Onboarding" {
		t.Errorf("name = %v, want Onboarding", data["name"])
	}
}

// --- Transcript ---

func TestTranscriptList(t *testing.T) {
	transcripts := []map[string]any{
		{
			"id": "t1", "sessionId": "s1", "moduleId": "m1", "moduleName": "Demo",
			"agentId": "a1", "userId": "u1", "partnerName": "Alice", "partnerCompany": "Acme",
			"createdAt": "2025-01-01T00:00:00Z", "messages": []any{},
		},
	}
	srv := httptest.NewServer(jsonHandler(map[string]any{
		"GET /transcripts": map[string]any{"transcripts": transcripts},
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "transcript", "list", "--api-url", srv.URL)
	if err != nil {
		t.Fatalf("transcript list error: %v", err)
	}
	if out == "" {
		t.Error("expected output")
	}
}

// --- Analytics ---

func TestAnalyticsSessions(t *testing.T) {
	resp := map[string]any{
		"stats": map[string]any{
			"totalSessions":         42,
			"avgSessionTimeMinutes": 5.5,
			"growth":                map[string]string{"sessions": "+10%", "avgTime": "+2%"},
		},
		"timeSeries": map[string]any{
			"sessions": []map[string]any{{"period": "2025-W01", "value": 10}},
		},
	}
	srv := httptest.NewServer(jsonHandler(map[string]any{
		"GET /analytics/sessions": resp,
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "analytics", "sessions", "--api-url", srv.URL)
	if err != nil {
		t.Fatalf("analytics sessions error: %v", err)
	}
	if out == "" {
		t.Error("expected output")
	}
}

func TestAnalyticsStages(t *testing.T) {
	resp := map[string]any{
		"stages":        []map[string]any{{"stage": "discovery", "count": 20, "percentage": 47.6}},
		"totalSessions": 42,
	}
	srv := httptest.NewServer(jsonHandler(map[string]any{
		"GET /analytics/stages": resp,
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "analytics", "stages", "--api-url", srv.URL)
	if err != nil {
		t.Fatalf("analytics stages error: %v", err)
	}
	if out == "" {
		t.Error("expected output")
	}
}

// --- Connect ---

func TestConnect(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/connection-details" {
			_ = json.NewEncoder(w).Encode(map[string]string{
				"serverUrl":        "wss://livekit.example.com",
				"roomName":         "room-123",
				"participantToken": "jwt_token_here",
				"participantName":  "user-abc",
			})
			return
		}
		w.WriteHeader(404)
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "connect", "demo-agent", "--api-url", srv.URL, "--format", "json")
	if err != nil {
		t.Fatalf("connect error: %v", err)
	}

	var data map[string]any
	if err := json.Unmarshal([]byte(out), &data); err != nil {
		t.Fatalf("JSON parse: %v", err)
	}
	if data["roomName"] != "room-123" {
		t.Errorf("roomName = %v, want room-123", data["roomName"])
	}
}

// --- Raw API ---

func TestRawAPI_Get(t *testing.T) {
	srv := httptest.NewServer(jsonHandler(map[string]any{
		"GET /health": map[string]string{"status": "ok"},
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "api", "GET", "/health", "--api-url", srv.URL)
	if err != nil {
		t.Fatalf("api GET error: %v", err)
	}

	var data map[string]any
	if err := json.Unmarshal([]byte(out), &data); err != nil {
		t.Fatalf("JSON parse: %v", err)
	}
	if data["status"] != "ok" {
		t.Errorf("status = %v, want ok", data["status"])
	}
}

func TestRawAPI_Post(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "POST" {
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			_ = json.NewEncoder(w).Encode(map[string]any{"received": body})
			return
		}
		w.WriteHeader(404)
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "api", "POST", "/test", "-d", `{"key":"value"}`, "--api-url", srv.URL)
	if err != nil {
		t.Fatalf("api POST error: %v", err)
	}

	var data map[string]any
	if err := json.Unmarshal([]byte(out), &data); err != nil {
		t.Fatalf("JSON parse: %v", err)
	}
	received, _ := data["received"].(map[string]any)
	if received["key"] != "value" {
		t.Errorf("received.key = %v, want value", received["key"])
	}
}

// --- Table output tests ---

func TestTableOutput_AgentList(t *testing.T) {
	agents := []map[string]any{
		{"id": 1, "publicId": "agt_abc", "name": "Alpha", "slug": "alpha", "status": "active"},
		{"id": 2, "publicId": "agt_def", "name": "Beta", "slug": "beta", "status": "inactive"},
	}
	srv := httptest.NewServer(jsonHandler(map[string]any{
		"GET /agents": map[string]any{"agents": agents},
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "agent", "list", "--api-url", srv.URL, "--format", "table")
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	// Table output should have headers and data rows.
	lines := bytes.Split([]byte(out), []byte("\n"))
	if len(lines) < 3 { // header + 2 data rows + trailing newline
		t.Errorf("expected at least 3 lines, got %d:\n%s", len(lines), out)
	}
}
