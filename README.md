# Anvil

Command-line interface and MCP server for the [Sable](https://withsable.com) AI voice agent platform.

## Quick Start (2 minutes)

```bash
# 1. Install
go install github.com/sable-inc/anvil/cmd/anvil@latest

# 2. Authenticate
anvil auth login --token svc_your_token_here

# 3. Set your default org
mkdir -p ~/.config/anvil
cat > ~/.config/anvil/config.yaml << 'EOF'
default_org: "your-org"
api_url: "https://api.withsable.com"
EOF

# 4. Set up shell completions (zsh)
echo 'source <(anvil completion zsh)' >> ~/.zshrc
source ~/.zshrc

# 5. Verify
anvil auth status
anvil agent list
```

## Install

### Option 1: Homebrew (recommended)

```bash
brew install sable-inc/tap/anvil
```

### Option 2: Go Install

```bash
go install github.com/sable-inc/anvil/cmd/anvil@latest
```

This puts the `anvil` binary in `$GOPATH/bin` (usually `~/go/bin`). Make sure that's in your `$PATH`:

```bash
# Add to your shell profile if not already there
export PATH="$HOME/go/bin:$PATH"
```

### Option 3: Download Binary (no Go required)

Download the latest release from [GitHub Releases](https://github.com/sable-inc/anvil/releases):

**macOS (Apple Silicon):**
```bash
curl -sL https://github.com/sable-inc/anvil/releases/latest/download/anvil_darwin_arm64.tar.gz | tar xz
sudo mv anvil /usr/local/bin/
```

**macOS (Intel):**
```bash
curl -sL https://github.com/sable-inc/anvil/releases/latest/download/anvil_darwin_amd64.tar.gz | tar xz
sudo mv anvil /usr/local/bin/
```

**Linux (amd64):**
```bash
curl -sL https://github.com/sable-inc/anvil/releases/latest/download/anvil_linux_amd64.tar.gz | tar xz
sudo mv anvil /usr/local/bin/
```

Or download manually from the [Releases page](https://github.com/sable-inc/anvil/releases).

### Option 4: Build from Source

```bash
git clone https://github.com/sable-inc/anvil.git
cd anvil
make build
sudo mv bin/anvil /usr/local/bin/
```

### Verify Installation

```bash
anvil version
# anvil v0.1.0 (commit: abc1234, built: 2026-02-14T..., darwin/arm64)
```

## Authentication

Anvil authenticates with sable-api using a service token:

```bash
# Store your token (one-time)
anvil auth login --token svc_your_token_here

# Verify it works
anvil auth whoami    # shows stored token info
anvil auth status    # verifies API connectivity
```

Credentials are stored in `~/.config/anvil/credentials.json` (file permissions: `0600`).

To get a service token, ask your team admin or generate one in the Sable Platform dashboard.

## Configuration

Create `~/.config/anvil/config.yaml`:

```yaml
default_org: "my-org"
api_url: "https://api.withsable.com"
format: "table"
```

Override any setting with flags:

```bash
anvil --org other-org --format json agent list
```

| Setting | Flag | Config Key | Description |
|---------|------|-----------|-------------|
| Organization | `--org` | `default_org` | Default org for all commands |
| API URL | `--api-url` | `api_url` | Sable API base URL |
| Output format | `--format` | `format` | `table`, `json`, or `yaml` |
| No color | `--no-color` | — | Disable colored output |
| Verbose | `--verbose` | — | Enable debug logging |

## Shell Completions

Anvil has dynamic tab completions that suggest real resource names from the live API.

### Setup (one-time, permanent)

**Zsh** (default on macOS):
```bash
echo 'source <(anvil completion zsh)' >> ~/.zshrc
source ~/.zshrc
```

**Bash:**
```bash
echo 'source <(anvil completion bash)' >> ~/.bashrc
source ~/.bashrc
```

**Fish:**
```bash
anvil completion fish > ~/.config/fish/completions/anvil.fish
```

**PowerShell:**
```powershell
anvil completion powershell >> $PROFILE
```

### What You Get

After setup, press `TAB` anywhere:

```
anvil <TAB>                      # all subcommands
anvil agent get <TAB>            # suggests agent slugs + IDs from API
anvil journey delete <TAB>       # suggests journey slugs + IDs
anvil kb sync <TAB>              # suggests KB item IDs with names
anvil config pull <TAB>          # suggests config version IDs
anvil connect <TAB>              # suggests agent slugs
anvil --format <TAB>             # suggests table, json, yaml
```

Completions fetch live data from the API using your stored token. If you're not authenticated, it gracefully shows no suggestions (no errors).

### FAQ

**Do I need to regenerate completions when anvil or my shell updates?**
No. The `source <(anvil completion ...)` pattern re-generates on every new shell session automatically. For Fish, re-run the file write command after updating anvil if new commands were added.

## Commands

### Agents

```
anvil agent list                     # List agents
anvil agent get <id|slug>            # Get agent details
anvil agent create --name "My Agent" # Create agent
anvil agent update <id> --name "New" # Update agent
anvil agent delete <id>              # Delete agent
```

### Journeys

```
anvil journey list                   # List journeys
anvil journey get <id>               # Get journey with moments/transitions
anvil journey create --name "Flow"   # Create journey
```

### Knowledge Base

```
anvil kb list                        # List knowledge base items
anvil kb get <id>                    # Get item details
anvil kb search "query"              # Semantic search
anvil kb import-url <url>            # Import a URL
anvil kb import-sitemap <url> --name "Docs"  # Import from sitemap
anvil kb sync <id>                   # Sync item to vector store
anvil kb sync-all                    # Sync all items
anvil kb crawl <id>                  # Re-crawl a URL item
anvil kb delete <id>                 # Delete item
anvil kb job <jobId>                 # Check sitemap import job status
```

### Config-as-Code

```
anvil config list                    # List config versions
anvil config get <id>                # Get config version details
anvil config pull <id> -o config.yaml  # Download config as YAML
anvil config push config.yaml        # Upload config from YAML
anvil config validate config.yaml    # Validate config locally
anvil config diff config.yaml --id <id>  # Diff local vs remote
```

### Analytics & Transcripts

```
anvil analytics sessions             # Session analytics
anvil analytics stages               # Stage funnel analytics
anvil transcript list                # List transcripts
anvil transcript view <id>           # View transcript messages
```

### Deployments

```
anvil deploy history --org org_xxx            # List deployment history
anvil deploy trigger --org org_xxx            # Trigger a deployment
anvil deploy trigger --org org_xxx --watch    # Trigger and poll until complete
anvil deploy rollback --org org_xxx           # Rollback deployment
anvil deploy create --org org_xxx             # Create initial deployment
anvil deploy delete --org org_xxx             # Delete deployment
anvil deploy update-secrets --org org_xxx     # Update deployment secrets
anvil deploy pin-forge --org org_xxx --forge-version v1.0  # Pin forge version
```

### LiveKit

```
# Sessions
anvil livekit sessions list --org org_xxx     # List active rooms
anvil livekit sessions get <room> --org org_xxx
anvil livekit sessions close <room> --org org_xxx

# Agent management
anvil livekit agent list --org org_xxx        # List agents
anvil livekit agent status --org org_xxx      # Agent status
anvil livekit agent versions --org org_xxx    # Agent versions
anvil livekit agent logs --org org_xxx        # Agent logs
anvil livekit agent restart --org org_xxx     # Restart agent
anvil livekit agent delete --org org_xxx      # Delete agent

# Agent secrets
anvil livekit agent secrets list --org org_xxx
anvil livekit agent secrets set --org org_xxx --secret KEY=VALUE
anvil livekit agent secrets delete <name> --org org_xxx
```

### Forge (Version Management)

```
anvil forge versions                  # List forge versions
anvil forge branches                  # List forge branches
anvil forge commits main              # List commits on branch
anvil forge validate main             # Validate a git ref
```

### Video Processing

```
anvil video generate-moment --video-url <url>         # Start moment generation
anvil video generate-moment --video-url <url> --watch  # Generate and poll
anvil video generate-journey --video-url <url>         # Start journey generation
anvil video job-status <jobId>                         # Check job status
```

### Utilities

```
anvil health                         # API health check
anvil health --db                    # Include database health
anvil connect <agent-slug>           # Get LiveKit connection details
anvil api GET /agents                # Raw API request
anvil api POST /agents -d '{"name":"test"}'
anvil version                        # Print version info
```

## Config-as-Code Workflow

Pull a config, edit locally, validate, and push:

```bash
# Pull the current config
anvil config pull <config-id> -o agent.yaml

# Edit the YAML file
$EDITOR agent.yaml

# Validate locally (no API call needed)
anvil config validate agent.yaml

# See what changed
anvil config diff agent.yaml --id <config-id>

# Push the updated config
anvil config push agent.yaml
```

## MCP Server (for AI Assistants)

Anvil includes an MCP server (28 Sable tools + 5 HyperDX observability tools) for AI coding assistants like Claude Code, Cursor, Windsurf, etc.

### One-Time Setup

```bash
anvil auth login --token svc_your_token     # Sable API credentials
anvil settings set-hyperdx hdx_your_key     # HyperDX API key (optional)
```

### Add to Claude Code

```bash
claude mcp add sable -s user -- anvil mcp serve
```

Or add to `.mcp.json` in your project root:

```json
{
  "mcpServers": {
    "sable": {
      "command": "anvil",
      "args": ["mcp", "serve"]
    }
  }
}
```

No secrets in the config -- anvil reads credentials from `~/.config/anvil/`.

### Cursor / Windsurf Setup

Same JSON format -- add to your editor's MCP settings file.

### Sable Tools (28)

| Domain | Tools |
|--------|-------|
| Agents | `list_agents`, `get_agent`, `create_agent`, `update_agent`, `delete_agent` |
| Journeys | `list_journeys`, `get_journey`, `create_journey`, `update_journey`, `delete_journey` |
| Knowledge Base | `list_knowledge_base`, `get_knowledge_base_item`, `search_knowledge_base`, `import_knowledge_base_url`, `delete_knowledge_base_item`, `sync_knowledge_base_item` |
| Config | `list_configs`, `get_config` |
| Deployments | `get_deploy_history`, `trigger_deploy`, `rollback_deploy` |
| Transcripts | `list_transcripts`, `get_transcript` |
| Analytics | `get_session_analytics`, `get_stage_analytics` |
| Utilities | `check_health`, `get_connection_details`, `raw_api_request` |

### HyperDX Observability Tools (5, optional)

Enabled when `anvil settings set-hyperdx <key>` has been run or `HYPERDX_API_KEY` env var is set.

| Tool | Purpose |
|------|---------|
| `hdx_search_events` | Query logs/spans with aggregations, time ranges, filters, group-by |
| `hdx_query_metrics` | Query metrics (Sum, Gauge, Histogram) |
| `hdx_list_dashboards` | List HyperDX dashboards |
| `hdx_get_dashboard` | Get dashboard with chart definitions |
| `hdx_list_alerts` | List configured alerts |

## Releasing New Versions

Anvil uses [goreleaser](https://goreleaser.com/) with GitHub Actions for automated releases.

### Creating a Release

```bash
git tag v0.1.0
git push origin v0.1.0
```

GitHub Actions will automatically:
1. Build binaries for Linux, macOS, and Windows (amd64 + arm64)
2. Create a GitHub Release with downloadable archives
3. Generate checksums and a changelog
4. Publish to the Homebrew tap (`sable-inc/homebrew-tap`)
5. Publish to the MCP Registry (discoverable in Claude Code / Cursor)

Team members can then upgrade via:
```bash
# Homebrew
brew upgrade anvil

# Go install
go install github.com/sable-inc/anvil/cmd/anvil@latest
```

### Snapshot Build (local testing)

```bash
goreleaser --snapshot --clean
ls dist/
```

## Development

```bash
make build     # Build binary
make test      # Run tests with race detector
make lint      # Run golangci-lint v2
make check     # Build + test + lint
make install   # Install to $GOPATH/bin
make generate  # Generate API client from OpenAPI spec
make clean     # Remove build artifacts
```

Requires:
- Go 1.26+
- golangci-lint v2 (for `make lint`)
- Running sable-api instance (for `make generate`)

## CI/CD

GitHub Actions runs automatically:
- **CI** (`.github/workflows/ci.yml`): build + test + lint on every push and PR
- **Release** (`.github/workflows/release.yml`): goreleaser on version tags (`v*`)
