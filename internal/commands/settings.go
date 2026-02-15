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
				_, _ = fmt.Fprintf(w, "  %-12s %s    # %s\n", key+":", val, desc)
			}

			show("api_url", cfg.APIURL, "Sable API base URL")
			show("default_org", cfg.DefaultOrg, "Default --org value")
			show("format", cfg.Format, "Output format (table|json|yaml)")
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

func newSettingsPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print the config file path",
		Run: func(cmd *cobra.Command, _ []string) {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), config.Path())
		},
	}
}
