package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/sable-inc/anvil/internal/config"
	"github.com/sable-inc/anvil/internal/hyperdx"
	"github.com/sable-inc/anvil/internal/mcp"
)

const defaultHyperDXURL = "https://api.hyperdx.io"

func newMCPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "MCP server for AI assistants",
		Long:  "Model Context Protocol server that exposes Sable Platform operations as tools for LLM assistants (Claude Code, Cursor, etc.).",
	}

	cmd.AddCommand(newMCPServeCmd())
	return cmd
}

func newMCPServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the MCP server over stdio",
		Long: `Starts a JSON-RPC 2.0 MCP server over stdin/stdout.

After one-time setup (anvil auth login + anvil settings set-hyperdx),
add to your Claude Code or Cursor MCP configuration:

  {
    "mcpServers": {
      "sable": {
        "command": "anvil",
        "args": ["mcp", "serve"]
      }
    }
  }

Or via CLI: claude mcp add sable -s user -- anvil mcp serve

Credentials are resolved from (highest priority first):
  1. Environment variables (ANVIL_TOKEN, ANVIL_API_URL, HYPERDX_API_KEY, HYPERDX_API_URL)
  2. Config file (~/.config/anvil/config.yaml, set via 'anvil settings set-hyperdx')
  3. Stored credentials (~/.config/anvil/credentials.json, set via 'anvil auth login')
  4. CLI flags (--org, --api-url)`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			a := AppFrom(cmd)

			a.Out = os.Stderr

			if envToken := os.Getenv("ANVIL_TOKEN"); envToken != "" && a.Token == "" {
				a.Token = envToken
			}
			if envURL := os.Getenv("ANVIL_API_URL"); envURL != "" && a.APIURL == "" {
				a.APIURL = envURL
			}

			client, err := a.RequireAuth()
			if err != nil {
				return fmt.Errorf("MCP server requires authentication: %w\nSet ANVIL_TOKEN env var or run 'anvil auth login' first", err)
			}

			var opts []mcp.ServerOption

			// Resolve HyperDX credentials: env vars > config file.
			hdxKey, hdxURL := resolveHyperDX()
			if hdxKey != "" {
				if hdxURL == "" {
					hdxURL = defaultHyperDXURL
				}
				opts = append(opts, mcp.WithHyperDX(hyperdx.NewClient(hdxURL, hdxKey)))
				_, _ = fmt.Fprintf(os.Stderr, "HyperDX tools enabled (%s)\n", hdxURL)
			}

			s := mcp.NewServer(client, a.OrgID, opts...)
			return mcp.Serve(s)
		},
	}
}

// resolveHyperDX returns the API key and URL, resolving each field independently:
// env var > config file. This matches how root.go resolves apiURL and orgFlag.
func resolveHyperDX() (apiKey, apiURL string) {
	apiKey = os.Getenv("HYPERDX_API_KEY")
	apiURL = os.Getenv("HYPERDX_API_URL")

	cfg, _ := config.Load()
	if cfg != nil {
		if apiKey == "" {
			apiKey = cfg.HyperDXAPIKey
		}
		if apiURL == "" {
			apiURL = cfg.HyperDXAPIURL
		}
	}
	return apiKey, apiURL
}
