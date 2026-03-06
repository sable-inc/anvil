package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/sable-inc/anvil/internal/config"
)

func newSettingsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "settings",
		Short: "Manage CLI settings",
		Long:  "View and update Anvil CLI settings stored in " + config.Path(),
	}

	cmd.AddCommand(newSettingsShowCmd())
	cmd.AddCommand(newSettingsSetCmd())
	cmd.AddCommand(newSettingsSetHyperDXCmd())
	cmd.AddCommand(newSettingsPathCmd())
	return cmd
}

func newSettingsShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current CLI settings",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			w := cmd.OutOrStdout()
			_, _ = fmt.Fprintf(w, "Config file: %s\n\n", config.Path())

			show := func(key, val, desc string) {
				if val == "" {
					val = "(not set)"
				}
				_, _ = fmt.Fprintf(w, "  %-18s %s    # %s\n", key+":", val, desc)
			}

			show("api_url", cfg.APIURL, "Sable API base URL")
			show("default_org", cfg.DefaultOrg, "Default --org value")
			show("format", cfg.Format, "Output format (table|json|yaml)")

			show("hyperdx_api_key", maskKey(cfg.HyperDXAPIKey), "HyperDX API key")
			show("hyperdx_api_url", cfg.HyperDXAPIURL, "HyperDX API URL")
			return nil
		},
	}
}

func newSettingsSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a CLI setting",
		Long: `Set a persistent CLI setting.

Available keys:
  api-url       Sable API base URL
  default-org   Default organization ID or slug
  format        Output format (table, json, yaml)`,
		Args:      cobra.ExactArgs(2),
		ValidArgs: []string{"api-url", "default-org", "format"},
		RunE: func(cmd *cobra.Command, args []string) error {
			key, value := args[0], args[1]

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			switch key {
			case "api-url":
				cfg.APIURL = value
			case "default-org":
				cfg.DefaultOrg = value
			case "format":
				switch value {
				case "table", "json", "yaml":
				default:
					return fmt.Errorf("invalid format %q (must be table, json, or yaml)", value)
				}
				cfg.Format = value
			default:
				return fmt.Errorf("unknown setting %q (valid: api-url, default-org, format)", key)
			}

			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Set %s = %s\n", key, value)
			return err
		},
	}
}

func newSettingsSetHyperDXCmd() *cobra.Command {
	var apiURL string

	cmd := &cobra.Command{
		Use:   "set-hyperdx <api-key>",
		Short: "Configure HyperDX API credentials",
		Long: `Stores a HyperDX personal API key for the MCP server.

After running this, the MCP server will automatically enable HyperDX
observability tools (hdx_search_events, hdx_query_metrics, etc.)
without requiring environment variables.

Get your API key from: https://www.hyperdx.io/team → API Keys`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			cfg.HyperDXAPIKey = args[0]
			if apiURL != "" {
				cfg.HyperDXAPIURL = apiURL
			}

			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}

			w := cmd.OutOrStdout()
			_, _ = fmt.Fprintf(w, "HyperDX API key saved to %s\n", config.Path())
			if apiURL != "" {
				_, _ = fmt.Fprintf(w, "HyperDX API URL: %s\n", apiURL)
			}
			_, _ = fmt.Fprintln(w, "MCP server will now include HyperDX tools on next start.")
			return nil
		},
	}

	cmd.Flags().StringVar(&apiURL, "api-url", "", "HyperDX API URL (default: https://api.hyperdx.io)")
	return cmd
}

const maskVisibleChars = 4

func maskKey(key string) string {
	if key == "" {
		return ""
	}
	if len(key) <= maskVisibleChars*2 {
		return "****"
	}
	return key[:maskVisibleChars] + "..." + key[len(key)-maskVisibleChars:]
}

func newSettingsPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print the config file path",
		Run: func(cmd *cobra.Command, _ []string) {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), config.Path())
		},
	}
}
