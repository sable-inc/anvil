# Anvil

Command-line interface and MCP server for the [Sable](https://withsable.com) AI voice agent platform.

## Install

```bash
# From source
go install github.com/sable-inc/anvil/cmd/anvil@latest

# Or build locally
git clone https://github.com/sable-inc/anvil.git
cd anvil
make build
./bin/anvil --help
```

## Authentication

Anvil authenticates with sable-api using a service token:

```bash
# Set your token
anvil auth login --token svc_your_token_here

# Verify
anvil auth whoami
```

Tokens are stored in `~/.config/anvil/credentials.json`.

## Configuration

Create `~/.config/anvil/config.yaml`:

```yaml
default_org: "my-org"
api_url: "https://api.withsable.com"
format: "table"
```

Override with flags:

```bash
anvil --org other-org --format json agent list
```

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

### Utilities

```
anvil health                         # API health check
anvil health --db                    # Include database health
anvil connect <agent-slug>           # Get LiveKit connection details
anvil api GET /agents                # Raw API request
anvil api POST /agents -d '{"name":"test"}'
anvil auth login --token <token>     # Store credentials
anvil auth logout                    # Clear credentials
anvil auth whoami                    # Show stored token
anvil auth status                    # Verify connectivity
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

## MCP Server

Anvil includes an MCP server for AI coding assistants (Claude Code, Cursor, etc.):

```json
{
  "mcpServers": {
    "anvil": {
      "command": "anvil",
      "args": ["mcp", "serve"]
    }
  }
}
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
