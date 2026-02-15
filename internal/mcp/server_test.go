package mcp_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/sable-inc/anvil/internal/api"
	sableMCP "github.com/sable-inc/anvil/internal/mcp"
)

// roundTripFunc implements http.RoundTripper for test mocking.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// mockClient returns an api.Client that uses the given handler for all requests.
func mockClient(handler func(*http.Request) (int, any)) *api.Client {
	doer := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			code, body := handler(req)
			data, _ := json.Marshal(body)
			return &http.Response{
				StatusCode: code,
				Body:       io.NopCloser(strings.NewReader(string(data))),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		}),
	}
	return api.NewClientWithDoer("http://test.local", doer)
}

func TestNewServer_RegistersTools(t *testing.T) {
	client := mockClient(func(_ *http.Request) (int, any) {
		return 200, map[string]any{}
	})

	s := sableMCP.NewServer(client, "42")

	resp := s.HandleMessage(context.Background(), json.RawMessage(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/list",
		"params": {}
	}`))

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}

	var result struct {
		Result struct {
			Tools []struct {
				Name string `json:"name"`
			} `json:"tools"`
		} `json:"result"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal tools/list response: %v", err)
	}

	expectedTools := []string{
		"list_agents", "get_agent", "create_agent", "update_agent", "delete_agent",
		"list_journeys", "get_journey", "create_journey", "update_journey", "delete_journey",
		"list_knowledge_base", "get_knowledge_base_item", "search_knowledge_base",
		"import_knowledge_base_url", "delete_knowledge_base_item", "sync_knowledge_base_item",
		"list_configs", "get_config",
		"get_deploy_history", "trigger_deploy", "rollback_deploy",
		"list_transcripts", "get_transcript",
		"get_session_analytics", "get_stage_analytics",
		"check_health", "get_connection_details", "raw_api_request",
	}

	toolNames := make(map[string]bool)
	for _, tool := range result.Result.Tools {
		toolNames[tool.Name] = true
	}

	for _, name := range expectedTools {
		if !toolNames[name] {
			t.Errorf("expected tool %q not found in tools/list response", name)
		}
	}

	if len(result.Result.Tools) != len(expectedTools) {
		t.Errorf("expected %d tools, got %d", len(expectedTools), len(result.Result.Tools))
	}
}

func TestListAgents_Success(t *testing.T) {
	agents := []map[string]any{
		{"id": 1, "name": "Agent One", "slug": "agent-one"},
		{"id": 2, "name": "Agent Two", "slug": "agent-two"},
	}
	client := mockClient(func(req *http.Request) (int, any) {
		if req.URL.Path == "/agents" && req.URL.Query().Get("orgId") == "42" {
			return 200, map[string]any{"agents": agents}
		}
		return 404, map[string]string{"error": "not found"}
	})

	s := sableMCP.NewServer(client, "42")
	result := callTool(t, s, "list_agents", map[string]any{})

	if result.IsError {
		t.Fatalf("expected success, got error: %s", textContent(result))
	}

	text := textContent(result)
	if !strings.Contains(text, "Agent One") || !strings.Contains(text, "Agent Two") {
		t.Errorf("expected agent names in response, got: %s", text)
	}
}

func TestGetAgent_NotFound(t *testing.T) {
	client := mockClient(func(_ *http.Request) (int, any) {
		return 404, map[string]string{"error": "Agent not found"}
	})

	s := sableMCP.NewServer(client, "42")
	result := callTool(t, s, "get_agent", map[string]any{"agent_id": "999"})

	if !result.IsError {
		t.Fatal("expected error result for not-found agent")
	}
	text := textContent(result)
	if !strings.Contains(text, "Failed to get agent") {
		t.Errorf("expected error message, got: %s", text)
	}
}

func TestCreateAgent_Success(t *testing.T) {
	client := mockClient(func(req *http.Request) (int, any) {
		if req.Method == "POST" && req.URL.Path == "/agents" {
			return 201, map[string]any{"agent": map[string]any{"id": 99, "name": "New Agent"}}
		}
		return 400, map[string]string{"error": "bad request"}
	})

	s := sableMCP.NewServer(client, "42")
	result := callTool(t, s, "create_agent", map[string]any{"name": "New Agent"})

	if result.IsError {
		t.Fatalf("expected success, got error: %s", textContent(result))
	}
	if !strings.Contains(textContent(result), "New Agent") {
		t.Errorf("expected agent name in response, got: %s", textContent(result))
	}
}

func TestSearchKB_Success(t *testing.T) {
	client := mockClient(func(req *http.Request) (int, any) {
		if req.Method == "POST" && req.URL.Path == "/knowledge-base/search" {
			return 200, map[string]any{
				"query":   "test query",
				"results": []map[string]any{{"id": "1", "score": 0.95, "content": "matching text"}},
			}
		}
		return 400, map[string]string{"error": "bad request"}
	})

	s := sableMCP.NewServer(client, "42")
	result := callTool(t, s, "search_knowledge_base", map[string]any{"query": "test query"})

	if result.IsError {
		t.Fatalf("expected success, got error: %s", textContent(result))
	}
	if !strings.Contains(textContent(result), "matching text") {
		t.Errorf("expected search result content, got: %s", textContent(result))
	}
}

func TestCheckHealth_Success(t *testing.T) {
	client := mockClient(func(req *http.Request) (int, any) {
		if req.URL.Path == "/health" {
			return 200, map[string]any{"status": "ok"}
		}
		return 500, map[string]string{"error": "internal"}
	})

	s := sableMCP.NewServer(client, "")
	result := callTool(t, s, "check_health", map[string]any{})

	if result.IsError {
		t.Fatalf("expected success, got error: %s", textContent(result))
	}
	if !strings.Contains(textContent(result), "ok") {
		t.Errorf("expected health status, got: %s", textContent(result))
	}
}

func TestRawAPIRequest_GET(t *testing.T) {
	client := mockClient(func(_ *http.Request) (int, any) {
		return 200, map[string]any{"data": "hello"}
	})

	s := sableMCP.NewServer(client, "")
	result := callTool(t, s, "raw_api_request", map[string]any{
		"method": "GET",
		"path":   "/health",
	})

	if result.IsError {
		t.Fatalf("expected success, got error: %s", textContent(result))
	}
}

func TestRawAPIRequest_InvalidBody(t *testing.T) {
	client := mockClient(func(_ *http.Request) (int, any) {
		return 200, map[string]any{}
	})

	s := sableMCP.NewServer(client, "")
	result := callTool(t, s, "raw_api_request", map[string]any{
		"method": "POST",
		"path":   "/test",
		"body":   "not valid json{",
	})

	if !result.IsError {
		t.Fatal("expected error for invalid JSON body")
	}
	if !strings.Contains(textContent(result), "Invalid JSON body") {
		t.Errorf("expected JSON body error, got: %s", textContent(result))
	}
}

func TestGetDeployHistory_RequiresOrgPublicID(t *testing.T) {
	client := mockClient(func(_ *http.Request) (int, any) {
		return 200, map[string]any{}
	})

	s := sableMCP.NewServer(client, "")
	result := callTool(t, s, "get_deploy_history", map[string]any{})

	if !result.IsError {
		t.Fatal("expected error when org_public_id is missing")
	}
	if !strings.Contains(textContent(result), "org_public_id is required") {
		t.Errorf("expected org_public_id error, got: %s", textContent(result))
	}
}

func TestOrgIDFallback(t *testing.T) {
	var capturedPath string
	client := mockClient(func(req *http.Request) (int, any) {
		capturedPath = req.URL.String()
		return 200, map[string]any{"agents": []any{}}
	})

	// Server configured with default orgID "42"
	s := sableMCP.NewServer(client, "42")
	result := callTool(t, s, "list_agents", map[string]any{})

	if result.IsError {
		t.Fatalf("expected success, got error: %s", textContent(result))
	}
	if !strings.Contains(capturedPath, "orgId=42") {
		t.Errorf("expected default org in request, got path: %s", capturedPath)
	}

	// Override org via parameter
	callTool(t, s, "list_agents", map[string]any{"org_id": "99"})
	if !strings.Contains(capturedPath, "orgId=99") {
		t.Errorf("expected overridden org in request, got path: %s", capturedPath)
	}
}

// --- Test helpers ---

// callTool invokes a tool on the MCP server using the JSON-RPC protocol.
func callTool(t *testing.T, s *server.MCPServer, name string, args map[string]any) *mcp.CallToolResult {
	t.Helper()

	argsJSON, _ := json.Marshal(args)
	reqJSON := fmt.Sprintf(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": %q,
			"arguments": %s
		}
	}`, name, string(argsJSON))

	resp := s.HandleMessage(context.Background(), json.RawMessage(reqJSON))

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}

	var wrapper struct {
		Result mcp.CallToolResult `json:"result"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		t.Fatalf("unmarshal tools/call response: %v\nraw: %s", err, string(data))
	}

	return &wrapper.Result
}

// textContent extracts the first text content from a tool result.
func textContent(result *mcp.CallToolResult) string {
	for _, c := range result.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			return tc.Text
		}
	}
	return ""
}
