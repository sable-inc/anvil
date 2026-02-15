package commands_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ─── Deploy ──────────────────────────────────────────────────────────────────

func TestDeployHistory(t *testing.T) {
	history := []map[string]any{
		{
			"id": "dep-001", "orgId": 28, "environment": "production",
			"forgeVersion": "v0.3.1", "status": "succeeded", "branch": "main",
			"createdAt": "2025-01-01T00:00:00Z", "updatedAt": "2025-01-01T00:00:00Z",
		},
		{
			"id": "dep-002", "orgId": 28, "environment": "test",
			"forgeVersion": nil, "status": "failed", "branch": nil,
			"createdAt": "2025-01-02T00:00:00Z", "updatedAt": "2025-01-02T00:00:00Z",
		},
	}
	srv := httptest.NewServer(jsonHandler(map[string]any{
		"GET /organizations/org_abc123/deploy-history": map[string]any{"deployHistory": history},
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "deploy", "history", "--org", "org_abc123", "--api-url", srv.URL)
	if err != nil {
		t.Fatalf("deploy history error: %v", err)
	}
	if out == "" {
		t.Error("expected output")
	}
}

func TestDeployHistory_JSON(t *testing.T) {
	history := []map[string]any{
		{"id": "dep-001", "orgId": 28, "environment": "production", "status": "succeeded", "createdAt": "2025-01-01T00:00:00Z", "updatedAt": "2025-01-01T00:00:00Z"},
	}
	srv := httptest.NewServer(jsonHandler(map[string]any{
		"GET /organizations/org_abc123/deploy-history": map[string]any{"deployHistory": history},
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "deploy", "history", "--org", "org_abc123", "--api-url", srv.URL, "--format", "json")
	if err != nil {
		t.Fatalf("deploy history --format json error: %v", err)
	}

	var data []map[string]any
	if err := json.Unmarshal([]byte(out), &data); err != nil {
		t.Fatalf("JSON parse: %v", err)
	}
	if len(data) != 1 {
		t.Errorf("expected 1 deployment, got %d", len(data))
	}
}

func TestDeployTrigger(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "POST" && r.URL.Path == "/organizations/org_abc123/deploy" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"deployment": map[string]any{
					"id": "dep-003", "orgId": 28, "environment": "production",
					"status": "pending", "createdAt": "2025-01-03T00:00:00Z", "updatedAt": "2025-01-03T00:00:00Z",
				},
				"message": "Deployment triggered",
			})
			return
		}
		w.WriteHeader(404)
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "deploy", "trigger", "--org", "org_abc123", "--api-url", srv.URL)
	if err != nil {
		t.Fatalf("deploy trigger error: %v", err)
	}
	if !strings.Contains(out, "Deployment triggered") {
		t.Errorf("expected 'Deployment triggered' in output, got: %s", out)
	}
}

func TestDeployRollback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "POST" && r.URL.Path == "/organizations/org_abc123/rollback" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"deployment": map[string]any{
					"id": "dep-004", "orgId": 28, "environment": "production",
					"status": "pending", "createdAt": "2025-01-04T00:00:00Z", "updatedAt": "2025-01-04T00:00:00Z",
				},
				"message": "Rollback triggered",
			})
			return
		}
		w.WriteHeader(404)
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "deploy", "rollback", "--org", "org_abc123", "--api-url", srv.URL)
	if err != nil {
		t.Fatalf("deploy rollback error: %v", err)
	}
	if !strings.Contains(out, "Rollback triggered") {
		t.Errorf("expected 'Rollback triggered' in output, got: %s", out)
	}
}

func TestDeployDelete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "DELETE" && r.URL.Path == "/organizations/org_abc123/deployment" {
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "Deployment deleted"})
			return
		}
		w.WriteHeader(404)
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "deploy", "delete", "--org", "org_abc123", "--api-url", srv.URL)
	if err != nil {
		t.Fatalf("deploy delete error: %v", err)
	}
	if !strings.Contains(out, "Deployment deleted") {
		t.Errorf("expected 'Deployment deleted' in output, got: %s", out)
	}
}

func TestDeployPinForge(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "PUT" && r.URL.Path == "/organizations/org_abc123/pin-forge-version" {
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "Pinned to v1.0.0 for production"})
			return
		}
		w.WriteHeader(404)
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "deploy", "pin-forge", "--org", "org_abc123", "--forge-version", "v1.0.0", "--api-url", srv.URL)
	if err != nil {
		t.Fatalf("deploy pin-forge error: %v", err)
	}
	if !strings.Contains(out, "Pinned to v1.0.0") {
		t.Errorf("expected pin message, got: %s", out)
	}
}

// Deploy with numeric orgId that resolves to publicId.
func TestDeployHistory_NumericOrg(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method + " " + r.URL.Path {
		case "GET /organizations/28":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"organization": map[string]any{"publicId": "org_abc123"},
			})
		case "GET /organizations/org_abc123/deploy-history":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"deployHistory": []map[string]any{
					{"id": "dep-001", "status": "succeeded", "createdAt": "2025-01-01T00:00:00Z", "updatedAt": "2025-01-01T00:00:00Z"},
				},
			})
		default:
			w.WriteHeader(404)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Not Found: " + r.URL.Path})
		}
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "deploy", "history", "--org", "28", "--api-url", srv.URL, "--format", "json")
	if err != nil {
		t.Fatalf("deploy history with numeric org error: %v", err)
	}

	var data []map[string]any
	if err := json.Unmarshal([]byte(out), &data); err != nil {
		t.Fatalf("JSON parse: %v", err)
	}
	if len(data) != 1 {
		t.Errorf("expected 1 deployment, got %d", len(data))
	}
}

// ─── LiveKit ─────────────────────────────────────────────────────────────────

func TestLivekitSessionsList(t *testing.T) {
	rooms := []map[string]any{
		{"name": "room-123", "sid": "RM_abc", "numParticipants": 2, "creationTime": 1706745600, "metadata": "{}"},
	}
	srv := httptest.NewServer(jsonHandler(map[string]any{
		"GET /organizations/org_abc123/livekit/sessions": map[string]any{"rooms": rooms},
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "livekit", "sessions", "list", "--org", "org_abc123", "--api-url", srv.URL)
	if err != nil {
		t.Fatalf("livekit sessions list error: %v", err)
	}
	if out == "" {
		t.Error("expected output")
	}
}

func TestLivekitSessionsGet(t *testing.T) {
	srv := httptest.NewServer(jsonHandler(map[string]any{
		"GET /organizations/org_abc123/livekit/sessions/room-123": map[string]any{
			"room":         map[string]any{"name": "room-123", "sid": "RM_abc", "numParticipants": 2, "creationTime": 1706745600},
			"participants": []map[string]any{{"identity": "user-1", "sid": "PA_1", "state": 1, "joinedAt": 1706745600, "name": "User 1"}},
		},
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "livekit", "sessions", "get", "room-123", "--org", "org_abc123", "--api-url", srv.URL, "--format", "json")
	if err != nil {
		t.Fatalf("livekit sessions get error: %v", err)
	}

	var data map[string]any
	if err := json.Unmarshal([]byte(out), &data); err != nil {
		t.Fatalf("JSON parse: %v", err)
	}
}

func TestLivekitAgentList(t *testing.T) {
	agents := []map[string]any{
		{"id": "agt-001", "name": "my-agent", "status": "active"},
	}
	srv := httptest.NewServer(jsonHandler(map[string]any{
		"GET /organizations/org_abc123/livekit/agents": map[string]any{"agents": agents},
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "livekit", "agent", "list", "--org", "org_abc123", "--api-url", srv.URL)
	if err != nil {
		t.Fatalf("livekit agent list error: %v", err)
	}
	if out == "" {
		t.Error("expected output")
	}
}

func TestLivekitAgentStatus(t *testing.T) {
	srv := httptest.NewServer(jsonHandler(map[string]any{
		"GET /organizations/org_abc123/livekit/agent/status": map[string]any{
			"status": map[string]any{
				"status": "active", "replicas": 2, "currentVersion": "v3",
				"uptime": "2d 5h", "image": "livekit/agent:v3",
			},
		},
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "livekit", "agent", "status", "--org", "org_abc123", "--api-url", srv.URL, "--format", "json")
	if err != nil {
		t.Fatalf("livekit agent status error: %v", err)
	}

	var data map[string]any
	if err := json.Unmarshal([]byte(out), &data); err != nil {
		t.Fatalf("JSON parse: %v", err)
	}
	if data["status"] != "active" {
		t.Errorf("status = %v, want active", data["status"])
	}
}

func TestLivekitAgentVersions(t *testing.T) {
	versions := []map[string]any{
		{"versionId": "v3", "createdAt": "2025-01-01", "status": "active", "image": "livekit/agent:v3"},
	}
	srv := httptest.NewServer(jsonHandler(map[string]any{
		"GET /organizations/org_abc123/livekit/agent/versions": map[string]any{"versions": versions},
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "livekit", "agent", "versions", "--org", "org_abc123", "--api-url", srv.URL)
	if err != nil {
		t.Fatalf("livekit agent versions error: %v", err)
	}
	if out == "" {
		t.Error("expected output")
	}
}

func TestLivekitAgentLogs(t *testing.T) {
	srv := httptest.NewServer(jsonHandler(map[string]any{
		"GET /organizations/org_abc123/livekit/agent/logs": map[string]any{
			"logs":    "2025-01-01 12:00:00 INFO Agent started\n",
			"logType": "deploy",
		},
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "livekit", "agent", "logs", "--org", "org_abc123", "--api-url", srv.URL)
	if err != nil {
		t.Fatalf("livekit agent logs error: %v", err)
	}
	if !strings.Contains(out, "Agent started") {
		t.Errorf("expected log output, got: %s", out)
	}
}

func TestLivekitAgentSecretsList(t *testing.T) {
	srv := httptest.NewServer(jsonHandler(map[string]any{
		"GET /organizations/org_abc123/livekit/agent/secrets": map[string]any{
			"secrets": []string{"OPENAI_API_KEY", "DATABASE_URL"},
		},
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "livekit", "agent", "secrets", "list", "--org", "org_abc123", "--api-url", srv.URL)
	if err != nil {
		t.Fatalf("livekit agent secrets list error: %v", err)
	}
	if out == "" {
		t.Error("expected output")
	}
}

func TestLivekitAgentSecretsSet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "POST" && r.URL.Path == "/organizations/org_abc123/livekit/agent/secrets" {
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "Secrets updated. Agent will perform a rolling restart."})
			return
		}
		w.WriteHeader(404)
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "livekit", "agent", "secrets", "set", "--org", "org_abc123",
		"--secret", "MY_KEY=my_value", "--api-url", srv.URL)
	if err != nil {
		t.Fatalf("livekit agent secrets set error: %v", err)
	}
	if !strings.Contains(out, "Secrets updated") {
		t.Errorf("expected 'Secrets updated' in output, got: %s", out)
	}
}

func TestLivekitAgentRestart(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "POST" && r.URL.Path == "/organizations/org_abc123/livekit/agent/restart" {
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "Agent restarted"})
			return
		}
		w.WriteHeader(404)
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "livekit", "agent", "restart", "--org", "org_abc123", "--api-url", srv.URL)
	if err != nil {
		t.Fatalf("livekit agent restart error: %v", err)
	}
	if !strings.Contains(out, "Agent restarted") {
		t.Errorf("expected 'Agent restarted' in output, got: %s", out)
	}
}

// ─── Forge ───────────────────────────────────────────────────────────────────

func TestForgeVersions(t *testing.T) {
	versions := []map[string]any{
		{"name": "v0.3.1", "sha": "abc123def456"},
		{"name": "v0.3.0", "sha": "111222333444"},
	}
	srv := httptest.NewServer(jsonHandler(map[string]any{
		"GET /forge-versions": map[string]any{"versions": versions},
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "forge", "versions", "--api-url", srv.URL)
	if err != nil {
		t.Fatalf("forge versions error: %v", err)
	}
	if out == "" {
		t.Error("expected output")
	}
}

func TestForgeVersions_JSON(t *testing.T) {
	versions := []map[string]any{
		{"name": "v0.3.1", "sha": "abc123def456"},
	}
	srv := httptest.NewServer(jsonHandler(map[string]any{
		"GET /forge-versions": map[string]any{"versions": versions},
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "forge", "versions", "--api-url", srv.URL, "--format", "json")
	if err != nil {
		t.Fatalf("forge versions --format json error: %v", err)
	}

	var data []map[string]any
	if err := json.Unmarshal([]byte(out), &data); err != nil {
		t.Fatalf("JSON parse: %v", err)
	}
	if len(data) != 1 {
		t.Errorf("expected 1 version, got %d", len(data))
	}
}

func TestForgeBranches(t *testing.T) {
	branches := []map[string]any{
		{"name": "main", "sha": "abc123"},
		{"name": "develop", "sha": "def456"},
	}
	srv := httptest.NewServer(jsonHandler(map[string]any{
		"GET /forge-branches": map[string]any{"branches": branches},
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "forge", "branches", "--api-url", srv.URL)
	if err != nil {
		t.Fatalf("forge branches error: %v", err)
	}
	if out == "" {
		t.Error("expected output")
	}
}

func TestForgeCommits(t *testing.T) {
	commits := []map[string]any{
		{"sha": "abc123def456789", "message": "feat: add voice timeout", "author": "John", "date": "2025-01-01T12:00:00Z"},
	}
	srv := httptest.NewServer(jsonHandler(map[string]any{
		"GET /forge-commits": map[string]any{"commits": commits},
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "forge", "commits", "main", "--api-url", srv.URL)
	if err != nil {
		t.Fatalf("forge commits error: %v", err)
	}
	if out == "" {
		t.Error("expected output")
	}
}

func TestForgeValidate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "POST" && r.URL.Path == "/forge-validate-ref" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"valid":       true,
				"resolvedSha": "abc123def456789012345678901234567890abcd",
			})
			return
		}
		w.WriteHeader(404)
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "forge", "validate", "main", "--api-url", srv.URL)
	if err != nil {
		t.Fatalf("forge validate error: %v", err)
	}
	if !strings.Contains(out, "Valid ref") {
		t.Errorf("expected 'Valid ref' in output, got: %s", out)
	}
}

func TestForgeValidate_Invalid(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "POST" && r.URL.Path == "/forge-validate-ref" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"valid": false,
				"error": "Ref not found",
			})
			return
		}
		w.WriteHeader(404)
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "forge", "validate", "nonexistent-branch", "--api-url", srv.URL)
	if err != nil {
		t.Fatalf("forge validate error: %v", err)
	}
	if !strings.Contains(out, "Invalid ref") {
		t.Errorf("expected 'Invalid ref' in output, got: %s", out)
	}
}

// ─── Video ───────────────────────────────────────────────────────────────────

func TestVideoJobStatus(t *testing.T) {
	srv := httptest.NewServer(jsonHandler(map[string]any{
		"GET /video-processing/jobs/job_123": map[string]any{
			"jobId":  "job_123",
			"status": "processing",
			"progress": map[string]any{
				"stage": "transcribing", "progress": 45, "message": "Transcribing audio...",
			},
		},
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "video", "job-status", "job_123", "--api-url", srv.URL, "--format", "json")
	if err != nil {
		t.Fatalf("video job-status error: %v", err)
	}

	var data map[string]any
	if err := json.Unmarshal([]byte(out), &data); err != nil {
		t.Fatalf("JSON parse: %v", err)
	}
	if data["status"] != "processing" {
		t.Errorf("status = %v, want processing", data["status"])
	}
}

func TestVideoJobStatus_Table(t *testing.T) {
	srv := httptest.NewServer(jsonHandler(map[string]any{
		"GET /video-processing/jobs/job_456": map[string]any{
			"jobId":  "job_456",
			"status": "completed",
			"result": map[string]any{"name": "Generated Moment"},
		},
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "video", "job-status", "job_456", "--api-url", srv.URL)
	if err != nil {
		t.Fatalf("video job-status table error: %v", err)
	}
	if out == "" {
		t.Error("expected output")
	}
}

func TestVideoGenerateMoment(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "POST" && r.URL.Path == "/video-processing/moment/start" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"jobId":  "job_new",
				"status": "pending",
			})
			return
		}
		w.WriteHeader(404)
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "video", "generate-moment", "--video-url", "https://example.com/video.mp4", "--api-url", srv.URL)
	if err != nil {
		t.Fatalf("video generate-moment error: %v", err)
	}
	if !strings.Contains(out, "Job started: job_new") {
		t.Errorf("expected 'Job started' in output, got: %s", out)
	}
}

func TestVideoGenerateJourney(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "POST" && r.URL.Path == "/video-processing/journey/start" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"jobId":  "job_journey",
				"status": "pending",
			})
			return
		}
		w.WriteHeader(404)
	}))
	defer srv.Close()

	testEnv(t)
	out, _, err := runCmd(t, "video", "generate-journey", "--video-url", "https://example.com/video.mp4", "--api-url", srv.URL)
	if err != nil {
		t.Fatalf("video generate-journey error: %v", err)
	}
	if !strings.Contains(out, "Job started: job_journey") {
		t.Errorf("expected 'Job started' in output, got: %s", out)
	}
}
