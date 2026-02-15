package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/sable-inc/anvil/internal/api"
	"github.com/sable-inc/anvil/internal/auth"
	"github.com/sable-inc/anvil/internal/output"
)

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
		Long:  "Login, logout, and check authentication status for the Sable API.",
	}

	cmd.AddCommand(newAuthLoginCmd())
	cmd.AddCommand(newAuthLogoutCmd())
	cmd.AddCommand(newAuthWhoamiCmd())
	cmd.AddCommand(newAuthStatusCmd())
	return cmd
}

func newAuthLoginCmd() *cobra.Command {
	var token string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with a service token",
		Long:  "Store a Sable API service token for CLI authentication.\nGet your token from the Sable Platform settings.",
		Example: `  # Login with a service token
  anvil auth login --token svc_your_token_here

  # Login by piping the token
  echo "svc_your_token_here" | anvil auth login --token -`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			tok := strings.TrimSpace(token)
			if tok == "" {
				return fmt.Errorf("--token is required")
			}

			if err := auth.Save(tok); err != nil {
				return fmt.Errorf("saving credentials: %w", err)
			}

			_, err := fmt.Fprintln(cmd.OutOrStdout(), "Logged in successfully. Credentials saved to", auth.Path())
			return err
		},
	}

	cmd.Flags().StringVar(&token, "token", "", "Service token (svc_...)")
	_ = cmd.MarkFlagRequired("token")
	return cmd
}

func newAuthLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove stored credentials",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := auth.Clear(); err != nil {
				return fmt.Errorf("clearing credentials: %w", err)
			}
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "Logged out. Credentials removed.")
			return err
		},
	}
}

func newAuthWhoamiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show stored credential info",
		Long:  "Display the currently stored authentication credentials.\nDoes not make an API call — use 'anvil auth status' to verify the token.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			creds, err := auth.Load()
			if err != nil {
				return fmt.Errorf("loading credentials: %w", err)
			}
			if creds == nil {
				return fmt.Errorf("not logged in — run 'anvil auth login' first")
			}

			a := AppFrom(cmd)
			f := output.New(a.Format)

			return f.Format(a.Out, map[string]string{
				"token":      maskToken(creds.Token),
				"created_at": creds.CreatedAt,
				"path":       auth.Path(),
			})
		},
	}
}

func newAuthStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Verify authentication against the API",
		Long:  "Check if your stored credentials are valid by calling the Sable API.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			a := AppFrom(cmd)

			client, err := a.RequireAuth()
			if err != nil {
				return err
			}

			// Hit /health to verify network connectivity, then /agents to verify auth.
			var health struct {
				Status string `json:"status"`
			}
			if err := client.Get(cmd.Context(), "/health", &health); err != nil {
				return fmt.Errorf("API unreachable: %w", err)
			}

			// /agents requires auth — if this fails with 401, the token is invalid.
			// A 400 "must provide orgId" means auth succeeded but the user is a Sable admin.
			var agents []any
			if err := client.Get(cmd.Context(), "/agents", &agents); err != nil {
				if _, ok := api.As[*api.UnauthorizedError](err); ok {
					return fmt.Errorf("token is invalid or expired — run 'anvil auth login' to re-authenticate")
				}
				if apiErr, ok := api.As[*api.ResponseError](err); ok && apiErr.StatusCode == 400 {
					// 400 means auth passed but endpoint needs orgId — token is valid.
					_, err = fmt.Fprintln(cmd.OutOrStdout(), "Authenticated. API is reachable and token is valid. (Sable admin — use --org to scope commands.)")
					return err
				}
				return fmt.Errorf("auth check failed: %w", err)
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Authenticated. API is reachable and token is valid.")
			return err
		},
	}
}

// maskToken shows only the prefix and last 4 characters of a token.
func maskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "..." + token[len(token)-4:]
}
