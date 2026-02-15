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

```
anvil agent list                     # List agents
anvil agent get <id>                 # Get agent details
anvil config pull <agent>            # Download agent config as YAML
anvil config push config.yaml        # Upload agent config
anvil config validate config.yaml    # Validate config locally
anvil config diff config.yaml        # Diff local vs remote
anvil journey list                   # List journeys
anvil kb list                        # List knowledge base items
anvil kb search "query"              # Search knowledge base
anvil deploy history                 # Deployment history
anvil deploy trigger                 # Trigger deployment
anvil health                         # API health check
anvil version                        # Print version info
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
make test      # Run tests
make lint      # Run linter
make generate  # Generate API client from OpenAPI spec
```

Requires:
- Go 1.24+
- golangci-lint v2 (for `make lint`)
- Running sable-api instance (for `make generate`)
