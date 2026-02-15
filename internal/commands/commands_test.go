package commands_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
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

// --- Knowledge Base ---

func TestKBList(t *testing.T) {
	items := []map[string]any{
		{"id": 1, "name": "Manual.pdf", "type": "document", "status": "ready", "chunkCount": 10, "createdAt": "2025-01-01T00:00:00Z", "updatedAt": "2025-01-01T00:00:00Z"},
		{"id": 2, "name": "FAQ Page", "type": "url", "status": "ready", "sourceUrl": "https://example.com/faq", "createdAt": "2025-01-01T00:00:00Z", "updatedAt": "2025-01-01T00:00:00Z"},
	}
	srv := httptest.NewServer(jsonHandler(map[string]any{
		"GET /knowledge-base": map[string]any{"items": items},
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "kb", "list", "--api-url", srv.URL)
	if err != nil {
		t.Fatalf("kb list error: %v", err)
	}
	if out == "" {
		t.Error("expected output")
	}
}

func TestKBGet(t *testing.T) {
	item := map[string]any{
		"id": 1, "name": "Manual.pdf", "type": "document", "status": "ready",
		"enabled": true, "chunkCount": 10, "wordCount": 5000,
		"createdAt": "2025-01-01T00:00:00Z", "updatedAt": "2025-01-01T00:00:00Z",
	}
	srv := httptest.NewServer(jsonHandler(map[string]any{
		"GET /knowledge-base/1": map[string]any{"item": item},
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "kb", "get", "1", "--api-url", srv.URL, "--format", "json")
	if err != nil {
		t.Fatalf("kb get error: %v", err)
	}

	var data map[string]any
	if err := json.Unmarshal([]byte(out), &data); err != nil {
		t.Fatalf("JSON parse: %v", err)
	}
	if data["name"] != "Manual.pdf" {
		t.Errorf("name = %v, want Manual.pdf", data["name"])
	}
}

func TestKBSearch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "POST" && r.URL.Path == "/knowledge-base/search" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"query": "test query",
				"results": []map[string]any{
					{"id": "kb-1-0", "score": 0.95, "content": "Result content", "metadata": map[string]any{}},
				},
			})
			return
		}
		w.WriteHeader(404)
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "kb", "search", "test query", "--api-url", srv.URL, "--format", "json")
	if err != nil {
		t.Fatalf("kb search error: %v", err)
	}

	var data map[string]any
	if err := json.Unmarshal([]byte(out), &data); err != nil {
		t.Fatalf("JSON parse: %v", err)
	}
	if data["query"] != "test query" {
		t.Errorf("query = %v, want 'test query'", data["query"])
	}
}

func TestKBImportURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "POST" && r.URL.Path == "/knowledge-base/url" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"item": map[string]any{
					"id": 99, "name": "Imported", "type": "url", "status": "pending",
					"sourceUrl": "https://example.com", "createdAt": "2025-01-01T00:00:00Z", "updatedAt": "2025-01-01T00:00:00Z",
				},
			})
			return
		}
		w.WriteHeader(404)
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "kb", "import-url", "https://example.com", "--api-url", srv.URL, "--format", "json")
	if err != nil {
		t.Fatalf("kb import-url error: %v", err)
	}

	var data map[string]any
	if err := json.Unmarshal([]byte(out), &data); err != nil {
		t.Fatalf("JSON parse: %v", err)
	}
	if data["name"] != "Imported" {
		t.Errorf("name = %v, want Imported", data["name"])
	}
}

func TestKBDelete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "DELETE" && r.URL.Path == "/knowledge-base/1" {
			_ = json.NewEncoder(w).Encode(map[string]any{"success": true})
			return
		}
		w.WriteHeader(404)
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "kb", "delete", "1", "--api-url", srv.URL)
	if err != nil {
		t.Fatalf("kb delete error: %v", err)
	}
	if out == "" {
		t.Error("expected output")
	}
}

func TestKBJob(t *testing.T) {
	srv := httptest.NewServer(jsonHandler(map[string]any{
		"GET /knowledge-base/jobs/job_123": map[string]any{
			"jobId":  "job_123",
			"status": "processing",
			"progress": map[string]any{
				"stage": "crawling", "progress": 45, "message": "Processing 50/100 URLs",
				"urlsDiscovered": 100, "urlsProcessed": 50,
			},
		},
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "kb", "job", "job_123", "--api-url", srv.URL, "--format", "json")
	if err != nil {
		t.Fatalf("kb job error: %v", err)
	}

	var data map[string]any
	if err := json.Unmarshal([]byte(out), &data); err != nil {
		t.Fatalf("JSON parse: %v", err)
	}
	if data["status"] != "processing" {
		t.Errorf("status = %v, want processing", data["status"])
	}
}

// --- Config ---

func TestConfigList(t *testing.T) {
	versions := []map[string]any{
		{"id": "uuid-1", "orgId": 1, "status": "draft", "createdAt": "2025-01-01T00:00:00Z", "updatedAt": "2025-01-01T00:00:00Z"},
	}
	srv := httptest.NewServer(jsonHandler(map[string]any{
		"GET /agent-configs": map[string]any{"configVersions": versions},
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "config", "list", "--api-url", srv.URL)
	if err != nil {
		t.Fatalf("config list error: %v", err)
	}
	if out == "" {
		t.Error("expected output")
	}
}

func TestConfigGet(t *testing.T) {
	cv := map[string]any{
		"id": "uuid-1", "orgId": 1, "status": "draft",
		"config":    map[string]any{"name": "Test Config"},
		"createdAt": "2025-01-01T00:00:00Z", "updatedAt": "2025-01-01T00:00:00Z",
	}
	srv := httptest.NewServer(jsonHandler(map[string]any{
		"GET /agent-configs/uuid-1": map[string]any{"configVersion": cv},
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "config", "get", "uuid-1", "--api-url", srv.URL, "--format", "json")
	if err != nil {
		t.Fatalf("config get error: %v", err)
	}

	var data map[string]any
	if err := json.Unmarshal([]byte(out), &data); err != nil {
		t.Fatalf("JSON parse: %v", err)
	}
	if data["status"] != "draft" {
		t.Errorf("status = %v, want draft", data["status"])
	}
}

func TestConfigValidate(t *testing.T) {
	// Write a valid YAML config to a temp file.
	cfgYAML := `config:
  name: "Test"
  greeting_instructions: "Hello"
  environment: "development"
  custom_tools: false
  llm:
    model: "gpt-4o-realtime-preview"
    modalities: ["text", "audio"]
    temperature: 0.7
    turn_detection:
      type: "server_vad"
      threshold: 0.5
      silence_duration_ms: 500
  stt:
    provider: "openai"
    openai:
      enabled: true
      model: "gpt-4o-mini-transcribe-2025-12-15"
    deepgram:
      enabled: false
      model: "nova-3"
      language: "multi"
      interim_results: false
      punctuate: false
      smart_format: false
  tts:
    provider: "none"
    voice_id: null
    model: null
  room:
    noise_cancellation: false
    video_enabled: false
    transcription_enabled: true
  components:
    browser:
      enabled: false
      enable_streaming: false
      mcp_server_dir: null
    vision:
      enabled: false
      proactive: false
      threshold: 0.5
      cooldown: 10
      debug_logs: false
    transcription:
      enabled: true
      utterance_finalize_delay: 1.0
      transcript_prefix: null
    transcription_logger:
      enabled: false
      module_id: null
      org_id: null
    memory:
      enabled: false
      summarization_model: "gpt-4o-mini"
    rag:
      enabled: false
      pinecone_index_name: null
      min_score: 0.5
      index_path: null
      rrf_k: 60
      search_top_n: 100
      search_k: 1000
      embeddings_model: "text-embedding-3-small"
      embeddings_dimension: 1536
      result_limit: 5
      no_results_message: "No results found."
    pip:
      enabled: false
`
	tmpDir := t.TempDir()
	cfgPath := tmpDir + "/config.yaml"
	if err := os.WriteFile(cfgPath, []byte(cfgYAML), 0o600); err != nil {
		t.Fatal(err)
	}

	out, _, err := runCmd(t, "config", "validate", cfgPath)
	if err != nil {
		t.Fatalf("config validate error: %v", err)
	}
	if out == "" {
		t.Error("expected output")
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
