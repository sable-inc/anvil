package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/sable-inc/anvil/internal/mcp"
)

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

Add to your Claude Code or Cursor MCP configuration:

  {
    "mcpServers": {
      "sable": {
        "command": "anvil",
        "args": ["mcp", "serve"],
        "env": {
          "ANVIL_API_URL": "https://api.withsable.com",
          "ANVIL_TOKEN": "svc_your_token"
        }
      }
    }
  }

The server uses credentials from:
  1. ANVIL_TOKEN / ANVIL_API_URL environment variables
  2. ~/.config/anvil/credentials.json and config.yaml (from 'anvil auth login')
  3. --org / --api-url flags on the parent command`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			a := AppFrom(cmd)

			// MCP servers MUST NOT write to stdout (it's the transport).
			// Redirect any verbose/error output to stderr.
			a.Out = os.Stderr

			// Allow env var overrides for token and API URL (common in MCP configs).
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

			s := mcp.NewServer(client, a.OrgID)
			return mcp.Serve(s)
		},
	}
}
