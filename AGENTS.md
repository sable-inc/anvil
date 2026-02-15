# AI Agent Documentation for Anvil CLI

This document provides context for AI agents working with the Anvil codebase.

## Scope

- The scope of this `AGENTS.md` is the `services/anvil/` directory.
- Direct instructions from prompts take priority over this file.
- For repository-wide context, see the root `CLAUDE.md`.

## Architecture Overview

Anvil is a Go CLI + MCP server for the Sable AI voice agent platform. It wraps the `sable-api` HTTP endpoints, providing command-line access to agents, configs, deployments, knowledge bases, and more.

```
cmd/
└── anvil/
    └── main.go              # Minimal entrypoint (<30 lines)

internal/
├── app/                     # Dependency container (App struct)
├── api/                     # HTTP client + error types
├── auth/                    # Credential storage + token refresh
├── commands/                # CLI command handlers (one file per domain)
├── config/                  # CLI config (~/.config/anvil/config.yaml)
├── configascode/            # YAML DSL for agent config (pull/push/validate/diff)
├── mcp/                     # MCP server (JSON-RPC 2.0 over stdio)
├── output/                  # Formatters (json, yaml, table)
└── version/                 # Build-time version info
```

## Go Conventions

### Interface-Driven Design

Interfaces are defined at the consumer site, not the implementor. Keep interfaces small (1-2 methods).

```go
// Defined where it's USED, not where it's implemented.
type Doer interface {
    Do(req *http.Request) (*http.Response, error)
}
```

### Constructor Injection

No DI framework, no globals. Dependencies flow through the `App` struct, created in `PersistentPreRunE` and stored in cobra's context.

```go
// Retrieve the App in any subcommand:
a := commands.AppFrom(cmd)
```

### Custom Error Types

Use typed errors with `errors.As()` for control flow, never sentinel errors.

```go
err := api.NewFromStatus(statusCode, message, hint)
if api.IsNotFound(err) {
    // handle 404
}
```

### Naming

| Type | Convention | Example |
|------|-----------|---------|
| Packages | lowercase, short | `api`, `config`, `output` |
| Interfaces | `-er` suffix for single-method | `Formatter`, `Doer` |
| Constructors | `New` + type name | `New()`, `NewFromStatus()` |
| Options | `With` prefix | `WithFormat()`, `WithVerbose()` |
| Files | snake_case | `errors.go`, `format_test.go` |

## Layer Responsibilities

| Layer | Responsibility | Imports From |
|-------|---------------|--------------|
| **cmd/anvil** | Entrypoint only | commands |
| **commands** | CLI wiring, flag parsing, call services | app, api, config, output |
| **app** | Dependency container | Nothing (leaf) |
| **api** | HTTP client, typed errors | Nothing (leaf) |
| **config** | Config file I/O | Nothing (leaf) |
| **output** | Format rendering | Nothing (leaf) |
| **version** | Build info | Nothing (leaf) |
| **auth** | Credential storage | config (for paths) |
| **configascode** | YAML DSL validation/conversion | Nothing (leaf) |
| **mcp** | MCP protocol | commands (reuses handlers) |

## Operational Commands (Phase 5)

### PublicID Resolution

Deploy and LiveKit endpoints use `publicId` (format: `org_xxx`) as path params, not the integer `orgId` used by CRUD commands. The `App.RequirePublicID(ctx)` helper handles both:
- Direct: `--org org_abc123` passes through
- Numeric: `--org 28` resolves via `GET /organizations/28` API lookup

### Deploy Commands

`deploy.go` — 7 subcommands for deployment management. All require `--org` with publicId.

| Command | API | Notes |
|---------|-----|-------|
| `deploy trigger` | `POST /organizations/{publicId}/deploy` | `--watch` polls until succeeded/failed |
| `deploy rollback` | `POST /organizations/{publicId}/rollback` | |
| `deploy create` | `POST /organizations/{publicId}/create-deployment` | |
| `deploy history` | `GET /organizations/{publicId}/deploy-history` | |
| `deploy delete` | `DELETE /organizations/{publicId}/deployment` | |
| `deploy update-secrets` | `POST /organizations/{publicId}/update-secrets` | |
| `deploy pin-forge` | `PUT /organizations/{publicId}/pin-forge-version` | `--forge-version` required |

### LiveKit Commands

`livekit.go` — 14 subcommands in nested structure. All require `--org` with publicId. Optional `--livekit-url`, `--livekit-api-key`, `--livekit-api-secret` credential overrides on parent command (sent as `x-livekit-*` headers via `*WithHeaders` client methods).

**Sessions group** (`anvil livekit sessions ...`):

| Command | API |
|---------|-----|
| `sessions list` | `GET /.../livekit/sessions` |
| `sessions get <room>` | `GET /.../livekit/sessions/{room}` |
| `sessions close <room>` | `DELETE /.../livekit/sessions/{room}` |
| `sessions remove-participant` | `DELETE /.../participants/{identity}` |
| `sessions mute` | `POST /.../participants/{identity}/mute` |

**Agent group** (`anvil livekit agent ...`):

| Command | API |
|---------|-----|
| `agent list` | `GET /.../livekit/agents` |
| `agent status` | `GET /.../livekit/agent/status` |
| `agent versions` | `GET /.../livekit/agent/versions` |
| `agent logs` | `GET /.../livekit/agent/logs` |
| `agent secrets list` | `GET /.../livekit/agent/secrets` |
| `agent secrets set` | `POST /.../livekit/agent/secrets` |
| `agent secrets delete` | `DELETE /.../livekit/agent/secrets/{name}` |
| `agent restart` | `POST /.../livekit/agent/restart` |
| `agent delete` | `DELETE /.../livekit/agent` |

### Forge Commands

`forge.go` — 4 subcommands for forge version management. Global admin endpoints (no `--org` needed).

| Command | API |
|---------|-----|
| `forge versions` | `GET /forge-versions` |
| `forge branches` | `GET /forge-branches` |
| `forge commits <branch>` | `GET /forge-commits?branch=<branch>` |
| `forge validate <ref>` | `POST /forge-validate-ref` |

### Video Commands

`video.go` — 3 subcommands for video processing. `--watch` on generate commands polls job status.

| Command | API |
|---------|-----|
| `video generate-moment` | `POST /video-processing/moment/start` |
| `video generate-journey` | `POST /video-processing/journey/start` |
| `video job-status <jobId>` | `GET /video-processing/jobs/{jobId}` |

### Async Polling (`output/spinner.go`)

Generic `Poll(ctx, w, PollConfig)` mechanism used by `deploy trigger --watch` and video `--watch` commands. Configurable interval, timeout, status callback.

## Config-as-Code System

The `configascode` package provides a local config management engine:

- **schema.go** — Go types mirroring the sable-api `AgentConfig` Zod schema (snake_case fields)
- **validate.go** — Local validation rules matching sable-api's `superRefine` cross-field checks
- **convert.go** — Bidirectional YAML <-> JSON conversion with `ConfigFile` wrapper type
- **diff.go** — Structured diff engine comparing local YAML against remote API config

### Validation Rules

The validator mirrors all 13 sable-api cross-field rules:
1. ElevenLabs TTS requires `voice_id` and `model`
2. None/OpenAI TTS must have null `voice_id` and `model`
3. STT provider must match enabled sub-providers
4. Vision proactive requires vision enabled
5. Browser streaming requires browser enabled
6. RAG enabled requires an index source
7. Embeddings dimension must match model

### Config File Format

```yaml
agent: my-agent          # Optional: agent slug
org_id: 42               # Optional: organization ID
config:
  name: "My Agent"
  environment: "production"
  # ... full AgentConfig fields
```

## MCP Server (Phase 6)

### Architecture

`internal/mcp/server.go` — Single-file MCP server using `mark3labs/mcp-go`. The `Handler` struct holds an `*api.Client` and default `orgID`, providing tool handlers that delegate to the sable-api.

**Transport**: stdio (JSON-RPC 2.0 over stdin/stdout). stdout is the transport — all logging goes to stderr.

**Tool naming**: `snake_case`, `verb_noun` pattern (e.g., `list_agents`, `create_journey`, `check_health`).

### Tools (28 total)

| Domain | Tools |
|--------|-------|
| **Agents** (5) | `list_agents`, `get_agent`, `create_agent`, `update_agent`, `delete_agent` |
| **Journeys** (5) | `list_journeys`, `get_journey`, `create_journey`, `update_journey`, `delete_journey` |
| **KB** (6) | `list_knowledge_base`, `get_knowledge_base_item`, `search_knowledge_base`, `import_knowledge_base_url`, `delete_knowledge_base_item`, `sync_knowledge_base_item` |
| **Config** (2) | `list_configs`, `get_config` |
| **Deploy** (3) | `get_deploy_history`, `trigger_deploy`, `rollback_deploy` |
| **Transcripts** (2) | `list_transcripts`, `get_transcript` |
| **Analytics** (2) | `get_session_analytics`, `get_stage_analytics` |
| **Utilities** (3) | `check_health`, `get_connection_details`, `raw_api_request` |

### Key patterns

- **JSON passthrough**: Handlers decode API responses as `json.RawMessage` and return as text content — no re-serialization.
- **Org context**: `optString(req, "org_id", h.orgID)` falls back to configured default. Mutation endpoints include `orgId` in request body via `setBodyOrgID()`. Read endpoints use `?orgId=X` query param via `withOrgQuery()`.
- **Error handling**: `errResult()` returns `isError: true` with actionable messages guiding the LLM to recovery tools.
- **Env var overrides**: `ANVIL_TOKEN` and `ANVIL_API_URL` environment variables for MCP config (since MCP clients can't pass CLI flags).

### MCP Client Configuration

```json
{
  "mcpServers": {
    "anvil": {
      "command": "anvil",
      "args": ["mcp", "serve"],
      "env": {
        "ANVIL_TOKEN": "svc_your_token_here",
        "ANVIL_API_URL": "http://localhost:8080"
      }
    }
  }
}
```

## Adding a New Command

1. Create `internal/commands/{domain}.go`
2. Define a constructor: `func newAgentCmd() *cobra.Command`
3. Use `AppFrom(cmd)` to get dependencies
4. Register in `root.go`: `root.AddCommand(newAgentCmd())`
5. Add tests in `{domain}_test.go`

```go
func newHealthCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "health",
        Short: "Check API health",
        RunE: func(cmd *cobra.Command, _ []string) error {
            a := AppFrom(cmd)
            // Use a.API, a.Out, a.Format, etc.
            return nil
        },
    }
}
```

## Error Handling

- Use `api.NewFromStatus()` to create typed errors from HTTP responses
- Use `api.IsNotFound()`, `api.IsUnauthorized()` etc. for type checks
- Commands return errors to cobra; cobra prints them via `SilenceErrors: true` pattern
- Never swallow errors silently

## Testing Patterns

- Table-driven tests with `t.Run()` for all non-trivial logic
- `_test` package suffix for black-box testing (e.g., `package api_test`)
- Use `bytes.Buffer` for capturing output in tests
- Run with race detector: `go test -race ./...`

## Quick Commands

```bash
make build     # Build binary to bin/anvil
make test      # Run tests with race detector
make lint      # golangci-lint v2
make install   # go install to $GOPATH/bin
make generate  # Fetch OpenAPI spec + generate client (requires running sable-api)
make clean     # Remove build artifacts
```

## Key Files to Reference

| File | Why |
|------|-----|
| `services/sable-api/src/routes/*.ts` | All API endpoints Anvil wraps |
| `services/sable-api/src/middleware/auth.ts` | Auth patterns (JWT, service token) |
| `services/sable-api/src/schemas/agent-config.ts` | Config validation rules |
| `services/sable-api/AGENTS.md` | API layer patterns |
