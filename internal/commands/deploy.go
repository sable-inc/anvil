package commands

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	"github.com/sable-inc/anvil/internal/api"
	"github.com/sable-inc/anvil/internal/output"
)

// Deployment mirrors the sable-api deploy history response shape.
type Deployment struct {
	ID                  string  `json:"id" yaml:"id"`
	OrgID               int     `json:"orgId" yaml:"orgId"`
	Environment         string  `json:"environment" yaml:"environment"`
	ForgeVersion        *string `json:"forgeVersion" yaml:"forgeVersion"`
	Status              string  `json:"status" yaml:"status"`
	LivekitBuildVersion *string `json:"livekitBuildVersion" yaml:"livekitBuildVersion"`
	TriggeredByAuthId   *string `json:"triggeredByAuthId" yaml:"triggeredByAuthId"`
	ErrorMessage        *string `json:"errorMessage" yaml:"errorMessage"`
	GithubRunId         *string `json:"githubRunId" yaml:"githubRunId"`
	CommitHash          *string `json:"commitHash" yaml:"commitHash"`
	Branch              *string `json:"branch" yaml:"branch"`
	CompletedAt         *string `json:"completedAt" yaml:"completedAt"`
	CreatedAt           string  `json:"createdAt" yaml:"createdAt"`
	UpdatedAt           string  `json:"updatedAt" yaml:"updatedAt"`
}

func newDeployCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Manage deployments",
		Long:  "Trigger, rollback, and inspect deployments for an organization.",
	}

	cmd.AddCommand(newDeployTriggerCmd())
	cmd.AddCommand(newDeployRollbackCmd())
	cmd.AddCommand(newDeployCreateCmd())
	cmd.AddCommand(newDeployHistoryCmd())
	cmd.AddCommand(newDeployDeleteCmd())
	cmd.AddCommand(newDeployUpdateSecretsCmd())
	cmd.AddCommand(newDeployPinForgeCmd())
	return cmd
}

func newDeployTriggerCmd() *cobra.Command {
	var (
		forgeVersion string
		environment  string
		branch       string
		watch        bool
	)

	cmd := &cobra.Command{
		Use:   "trigger",
		Short: "Trigger a deployment",
		RunE: func(cmd *cobra.Command, _ []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}
			publicID, err := a.RequirePublicID(cmd.Context())
			if err != nil {
				return err
			}

			body := map[string]any{}
			if forgeVersion != "" {
				body["forgeVersion"] = forgeVersion
			}
			if environment != "" {
				body["environment"] = environment
			}
			if branch != "" {
				body["branch"] = branch
			}

			var resp struct {
				Deployment Deployment `json:"deployment"`
				Message    string     `json:"message"`
			}
			path := "/organizations/" + publicID + "/deploy"
			if err := client.Post(cmd.Context(), path, body, &resp); err != nil {
				return err
			}

			d := resp.Deployment
			_, _ = fmt.Fprintf(a.Out, "Deployment triggered: %s (status: %s)\n", d.ID, d.Status)

			if !watch {
				return nil
			}

			return pollDeployStatus(cmd.Context(), a.Out, client, publicID, d.ID, environment)
		},
	}

	cmd.Flags().StringVar(&forgeVersion, "forge-version", "", "Forge version to deploy")
	cmd.Flags().StringVar(&environment, "environment", "", "Environment (production|test)")
	cmd.Flags().StringVar(&branch, "branch", "", "Git branch to deploy")
	cmd.Flags().BoolVar(&watch, "watch", false, "Poll until deployment completes")
	return cmd
}

func newDeployRollbackCmd() *cobra.Command {
	var (
		environment string
		watch       bool
	)

	cmd := &cobra.Command{
		Use:   "rollback",
		Short: "Rollback a deployment",
		RunE: func(cmd *cobra.Command, _ []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}
			publicID, err := a.RequirePublicID(cmd.Context())
			if err != nil {
				return err
			}

			body := map[string]any{}
			if environment != "" {
				body["environment"] = environment
			}

			var resp struct {
				Deployment Deployment `json:"deployment"`
				Message    string     `json:"message"`
			}
			path := "/organizations/" + publicID + "/rollback"
			if err := client.Post(cmd.Context(), path, body, &resp); err != nil {
				return err
			}

			d := resp.Deployment
			_, _ = fmt.Fprintf(a.Out, "Rollback triggered: %s (status: %s)\n", d.ID, d.Status)

			if !watch {
				return nil
			}

			return pollDeployStatus(cmd.Context(), a.Out, client, publicID, d.ID, environment)
		},
	}

	cmd.Flags().StringVar(&environment, "environment", "", "Environment (production|test)")
	cmd.Flags().BoolVar(&watch, "watch", false, "Poll until rollback completes")
	return cmd
}

func newDeployCreateCmd() *cobra.Command {
	var (
		forgeVersion string
		environment  string
		branch       string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create initial deployment for an org",
		RunE: func(cmd *cobra.Command, _ []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}
			publicID, err := a.RequirePublicID(cmd.Context())
			if err != nil {
				return err
			}

			body := map[string]any{}
			if forgeVersion != "" {
				body["forgeVersion"] = forgeVersion
			}
			if environment != "" {
				body["environment"] = environment
			}
			if branch != "" {
				body["branch"] = branch
			}

			var resp struct {
				Deployment Deployment `json:"deployment"`
				Message    string     `json:"message"`
			}
			path := "/organizations/" + publicID + "/create-deployment"
			if err := client.Post(cmd.Context(), path, body, &resp); err != nil {
				return err
			}

			d := resp.Deployment
			_, _ = fmt.Fprintf(a.Out, "Deployment created: %s (status: %s)\n", d.ID, d.Status)
			return nil
		},
	}

	cmd.Flags().StringVar(&forgeVersion, "forge-version", "", "Forge version")
	cmd.Flags().StringVar(&environment, "environment", "", "Environment (production|test)")
	cmd.Flags().StringVar(&branch, "branch", "", "Git branch")
	return cmd
}

func newDeployHistoryCmd() *cobra.Command {
	var environment string

	cmd := &cobra.Command{
		Use:   "history",
		Short: "Show deployment history",
		RunE: func(cmd *cobra.Command, _ []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}
			publicID, err := a.RequirePublicID(cmd.Context())
			if err != nil {
				return err
			}

			path := "/organizations/" + publicID + "/deploy-history"
			if environment != "" {
				path += "?environment=" + environment
			}

			var resp struct {
				DeployHistory []Deployment `json:"deployHistory"`
			}
			if err := client.Get(cmd.Context(), path, &resp); err != nil {
				return err
			}

			t := output.NewTable("ID", "Environment", "Forge Version", "Status", "Branch", "Created")
			for _, d := range resp.DeployHistory {
				t.AddRow(
					truncateID(d.ID),
					d.Environment,
					ptrStr(d.ForgeVersion),
					d.Status,
					ptrStr(d.Branch),
					d.CreatedAt,
				)
			}
			return output.Write(a.Out, a.Format, resp.DeployHistory, t)
		},
	}

	cmd.Flags().StringVar(&environment, "environment", "", "Filter by environment (production|test)")
	return cmd
}

func newDeployDeleteCmd() *cobra.Command {
	var environment string

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a deployment",
		RunE: func(cmd *cobra.Command, _ []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}
			publicID, err := a.RequirePublicID(cmd.Context())
			if err != nil {
				return err
			}

			path := "/organizations/" + publicID + "/deployment"
			if environment != "" {
				path += "?environment=" + environment
			}

			var resp struct {
				Message string `json:"message"`
			}
			if err := client.Delete(cmd.Context(), path, &resp); err != nil {
				return err
			}

			_, err = fmt.Fprintln(a.Out, resp.Message)
			return err
		},
	}

	cmd.Flags().StringVar(&environment, "environment", "", "Environment (production|test)")
	return cmd
}

func newDeployUpdateSecretsCmd() *cobra.Command {
	var environment string

	cmd := &cobra.Command{
		Use:   "update-secrets",
		Short: "Trigger secrets update for a deployment",
		RunE: func(cmd *cobra.Command, _ []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}
			publicID, err := a.RequirePublicID(cmd.Context())
			if err != nil {
				return err
			}

			body := map[string]any{}
			if environment != "" {
				body["environment"] = environment
			}

			var resp struct {
				Message string `json:"message"`
			}
			path := "/organizations/" + publicID + "/update-secrets"
			if err := client.Post(cmd.Context(), path, body, &resp); err != nil {
				return err
			}

			_, err = fmt.Fprintln(a.Out, resp.Message)
			return err
		},
	}

	cmd.Flags().StringVar(&environment, "environment", "", "Environment (production|test)")
	return cmd
}

func newDeployPinForgeCmd() *cobra.Command {
	var (
		forgeVersion string
		environment  string
	)

	cmd := &cobra.Command{
		Use:   "pin-forge",
		Short: "Pin or unpin a forge version",
		Long:  "Pin an org's deployment to a specific forge version, or pass --forge-version=\"\" to unpin.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}
			publicID, err := a.RequirePublicID(cmd.Context())
			if err != nil {
				return err
			}

			body := map[string]any{
				"forgeVersion": forgeVersion,
			}
			if environment != "" {
				body["environment"] = environment
			}

			var resp struct {
				Message string `json:"message"`
			}
			path := "/organizations/" + publicID + "/pin-forge-version"
			if err := client.Put(cmd.Context(), path, body, &resp); err != nil {
				return err
			}

			_, err = fmt.Fprintln(a.Out, resp.Message)
			return err
		},
	}

	cmd.Flags().StringVar(&forgeVersion, "forge-version", "", "Forge version to pin (empty to unpin)")
	cmd.Flags().StringVar(&environment, "environment", "", "Environment (production|test)")
	_ = cmd.MarkFlagRequired("forge-version")
	return cmd
}

// pollDeployStatus polls deploy history until the given deployment reaches a terminal state.
func pollDeployStatus(ctx context.Context, w io.Writer, client *api.Client, publicID, deployID, environment string) error {
	var lastStatus string
	return output.Poll(ctx, w, output.PollConfig{
		Interval: 3 * time.Second,
		Timeout:  10 * time.Minute,
		StatusFunc: func(ctx context.Context) (string, bool, error) {
			path := "/organizations/" + publicID + "/deploy-history"
			if environment != "" {
				path += "?environment=" + environment
			}
			var resp struct {
				DeployHistory []Deployment `json:"deployHistory"`
			}
			if err := client.Get(ctx, path, &resp); err != nil {
				return "", false, err
			}
			for _, d := range resp.DeployHistory {
				if d.ID == deployID {
					done := d.Status == "succeeded" || d.Status == "failed" || d.Status == "rolled_back"
					return d.Status, done, nil
				}
			}
			return "unknown", false, nil
		},
		OnStatus: func(status string) {
			if status != lastStatus {
				_, _ = fmt.Fprintf(w, "  status: %s\n", status)
				lastStatus = status
			}
		},
	})
}

// ptrStr dereferences a string pointer, returning "-" if nil.
func ptrStr(s *string) string {
	if s == nil {
		return "-"
	}
	return *s
}

// truncateID returns first 8 characters of an ID for compact table display.
func truncateID(id string) string {
	if len(id) > 12 {
		return id[:12] + "..."
	}
	return id
}
