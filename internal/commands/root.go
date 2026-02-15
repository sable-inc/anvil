// Package commands defines all CLI commands for the Anvil CLI.
package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/sable-inc/anvil/internal/app"
	"github.com/sable-inc/anvil/internal/auth"
	"github.com/sable-inc/anvil/internal/config"
)

type contextKey struct{}

// AppFrom retrieves the App from a command's context.
// Subcommands use this to access shared dependencies.
func AppFrom(cmd *cobra.Command) *app.App {
	a, _ := cmd.Context().Value(contextKey{}).(*app.App)
	return a
}

// NewRoot creates the root cobra command with global persistent flags.
func NewRoot() *cobra.Command {
	var (
		orgFlag string
		apiURL  string
		format  string
		noColor bool
		verbose bool
	)

	root := &cobra.Command{
		Use:   "anvil",
		Short: "Sable Platform CLI",
		Long:  "Anvil is the command-line interface for the Sable AI voice agent platform.\nManage agents, configs, deployments, knowledge bases, and more.",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// Skip wiring for completion and help commands.
			if cmd.Name() == "completion" || cmd.Name() == "help" {
				return nil
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			var opts []app.Option

			// Resolve org: flag > config default.
			org := orgFlag
			if org == "" {
				org = cfg.DefaultOrg
			}
			if org != "" {
				opts = append(opts, app.WithOrgID(org))
			}

			// Resolve API URL: flag > config.
			u := apiURL
			if u == "" {
				u = cfg.APIURL
			}
			if u != "" {
				opts = append(opts, app.WithAPIURL(u))
			}

			// Resolve format: flag > config > default.
			f := format
			if f == "" {
				f = cfg.Format
			}
			if f != "" {
				opts = append(opts, app.WithFormat(f))
			}

			// Wire App I/O to cobra's output writers so tests can capture output.
			opts = append(opts,
				app.WithOutput(cmd.OutOrStdout()),
				app.WithErrOutput(cmd.OutOrStderr()),
				app.WithVerbose(verbose),
				app.WithNoColor(noColor),
			)

			// Load stored credentials (optional — auth login doesn't need them).
			creds, credErr := auth.Load()
			if credErr != nil && verbose {
				_, _ = fmt.Fprintf(os.Stderr, "warning: could not load credentials: %v\n", credErr)
			}
			if creds != nil {
				opts = append(opts, app.WithToken(creds.Token))
			}

			a := app.New(opts...)
			ctx := context.WithValue(cmd.Context(), contextKey{}, a)
			cmd.SetContext(ctx)
			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	flags := root.PersistentFlags()
	flags.StringVar(&orgFlag, "org", "", "Organization ID or slug")
	flags.StringVar(&apiURL, "api-url", "", "Sable API base URL")
	flags.StringVar(&format, "format", "", "Output format: json, yaml, table (default: table)")
	flags.BoolVar(&noColor, "no-color", false, "Disable colored output")
	flags.BoolVar(&verbose, "verbose", false, "Enable verbose logging")

	root.AddCommand(newVersionCmd())
	root.AddCommand(newAuthCmd())
	root.AddCommand(newHealthCmd())
	root.AddCommand(newAgentCmd())
	root.AddCommand(newJourneyCmd())
	root.AddCommand(newTranscriptCmd())
	root.AddCommand(newAnalyticsCmd())
	root.AddCommand(newConnectCmd())
	root.AddCommand(newKBCmd())
	root.AddCommand(newConfigCmd())
	root.AddCommand(newRawAPICmd())

	return root
}

// Execute runs the root command. Called from main.
func Execute() {
	root := NewRoot()
	if err := root.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
