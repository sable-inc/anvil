package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/sable-inc/anvil/internal/configascode"
	"github.com/sable-inc/anvil/internal/output"
)

// AgentConfigVersion mirrors the sable-api AgentConfigVersion response shape.
type AgentConfigVersion struct {
	ID          string `json:"id" yaml:"id"`
	OrgID       int    `json:"orgId" yaml:"orgId"`
	AgentID     *int   `json:"agentId" yaml:"agentId"`
	Status      string `json:"status" yaml:"status"`
	Config      any    `json:"config" yaml:"config"`
	CreatedAt   string `json:"createdAt" yaml:"createdAt"`
	UpdatedAt   string `json:"updatedAt" yaml:"updatedAt"`
	PublishedAt *string `json:"publishedAt" yaml:"publishedAt"`
	ExpiresAt   *string `json:"expiresAt" yaml:"expiresAt"`
}

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "config",
		Aliases: []string{"configs"},
		Short:   "Manage agent configurations (config-as-code)",
	}

	cmd.AddCommand(newConfigListCmd())
	cmd.AddCommand(newConfigGetCmd())
	cmd.AddCommand(newConfigPullCmd())
	cmd.AddCommand(newConfigPushCmd())
	cmd.AddCommand(newConfigValidateCmd())
	cmd.AddCommand(newConfigDiffCmd())
	return cmd
}

func newConfigListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List agent config versions",
		RunE: func(cmd *cobra.Command, _ []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}

			path := "/agent-configs"
			if a.OrgID != "" {
				path += "?orgId=" + a.OrgID
			}

			var resp struct {
				ConfigVersions []AgentConfigVersion `json:"configVersions"`
			}
			if err := client.Get(cmd.Context(), path, &resp); err != nil {
				return err
			}

			t := output.NewTable("ID", "Agent ID", "Status", "Created", "Published")
			for _, cv := range resp.ConfigVersions {
				agentID := "-"
				if cv.AgentID != nil {
					agentID = strconv.Itoa(*cv.AgentID)
				}
				published := "-"
				if cv.PublishedAt != nil {
					published = *cv.PublishedAt
				}
				t.AddRow(cv.ID, agentID, cv.Status, cv.CreatedAt, published)
			}
			return output.Write(a.Out, a.Format, resp.ConfigVersions, t)
		},
	}
}

func newConfigGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <config-id>",
		Short: "Get an agent config version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}

			var resp struct {
				ConfigVersion AgentConfigVersion `json:"configVersion"`
			}
			if err := client.Get(cmd.Context(), "/agent-configs/"+args[0], &resp); err != nil {
				return err
			}

			cv := resp.ConfigVersion
			t := output.NewTable("Field", "Value")
			t.AddRow("ID", cv.ID)
			t.AddRow("Org ID", strconv.Itoa(cv.OrgID))
			agentID := "-"
			if cv.AgentID != nil {
				agentID = strconv.Itoa(*cv.AgentID)
			}
			t.AddRow("Agent ID", agentID)
			t.AddRow("Status", cv.Status)
			t.AddRow("Created", cv.CreatedAt)
			t.AddRow("Updated", cv.UpdatedAt)
			published := "-"
			if cv.PublishedAt != nil {
				published = *cv.PublishedAt
			}
			t.AddRow("Published", published)
			return output.Write(a.Out, a.Format, cv, t)
		},
	}
}

func newConfigPullCmd() *cobra.Command {
	var outFile string

	cmd := &cobra.Command{
		Use:   "pull <config-id>",
		Short: "Download an agent config as YAML",
		Long:  "Fetches a config version from the API and outputs it as a YAML file.\nUse --output to write to a file, or pipe stdout.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}

			var resp struct {
				ConfigVersion AgentConfigVersion `json:"configVersion"`
			}
			if err := client.Get(cmd.Context(), "/agent-configs/"+args[0], &resp); err != nil {
				return err
			}

			if resp.ConfigVersion.Config == nil {
				return fmt.Errorf("config version %s has no config data", args[0])
			}

			// Convert the config from API JSON to our typed struct.
			configJSON, err := json.Marshal(resp.ConfigVersion.Config)
			if err != nil {
				return fmt.Errorf("marshaling config: %w", err)
			}
			agentCfg, err := configascode.FromJSON(configJSON)
			if err != nil {
				return fmt.Errorf("parsing config: %w", err)
			}

			cfgFile := &configascode.ConfigFile{
				OrgID:  resp.ConfigVersion.OrgID,
				Config: *agentCfg,
			}

			if outFile != "" {
				data, err := configascode.ToYAML(cfgFile)
				if err != nil {
					return err
				}
				if err := os.WriteFile(outFile, data, 0o644); err != nil { //nolint:gosec // user writes their own config
					return fmt.Errorf("writing file: %w", err)
				}
				_, err = fmt.Fprintf(a.Out, "Config written to %s\n", outFile)
				return err
			}

			return configascode.WriteYAML(a.Out, cfgFile)
		},
	}

	cmd.Flags().StringVarP(&outFile, "output", "o", "", "Output file path (default: stdout)")
	return cmd
}

func newConfigPushCmd() *cobra.Command {
	var expiresAt string

	cmd := &cobra.Command{
		Use:   "push <config.yaml>",
		Short: "Upload an agent config from a YAML file",
		Long:  "Reads a YAML config file, validates it locally, then creates a new config version via the API.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}

			cfgFile, err := configascode.FromYAMLFile(args[0])
			if err != nil {
				return err
			}

			// Validate locally first.
			if vErr := configascode.Validate(&cfgFile.Config); vErr != nil {
				return fmt.Errorf("validation failed:\n%w", vErr)
			}

			body := map[string]any{
				"config": cfgFile.Config,
			}
			if cfgFile.OrgID > 0 {
				body["orgId"] = cfgFile.OrgID
			}
			if expiresAt != "" {
				body["expiresAt"] = expiresAt
			}

			var resp struct {
				ConfigVersion AgentConfigVersion `json:"configVersion"`
			}
			if err := client.Post(cmd.Context(), "/agent-configs", body, &resp); err != nil {
				return err
			}

			cv := resp.ConfigVersion
			_, err = fmt.Fprintf(a.Out, "Config version %s created (status: %s)\n", cv.ID, cv.Status)
			return err
		},
	}

	cmd.Flags().StringVar(&expiresAt, "expires-at", "", "Expiration timestamp (ISO 8601)")
	return cmd
}

func newConfigValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate <config.yaml>",
		Short: "Validate a config file locally",
		Long:  "Parses and validates a YAML config file against the agent config schema rules without contacting the API.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a := AppFrom(cmd)

			cfgFile, err := configascode.FromYAMLFile(args[0])
			if err != nil {
				return err
			}

			if vErr := configascode.Validate(&cfgFile.Config); vErr != nil {
				return fmt.Errorf("validation failed:\n%w", vErr)
			}

			_, err = fmt.Fprintf(a.Out, "Config %s is valid.\n", args[0])
			return err
		},
	}
}

func newConfigDiffCmd() *cobra.Command {
	var configID string

	cmd := &cobra.Command{
		Use:   "diff <config.yaml>",
		Short: "Diff local config against a remote config version",
		Long:  "Compares a local YAML config file against a remote config version and shows the differences.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}

			if configID == "" {
				return fmt.Errorf("--id is required (specify the remote config version ID to compare against)")
			}

			// Load local config.
			cfgFile, err := configascode.FromYAMLFile(args[0])
			if err != nil {
				return err
			}

			// Fetch remote config.
			var resp struct {
				ConfigVersion AgentConfigVersion `json:"configVersion"`
			}
			if err := client.Get(cmd.Context(), "/agent-configs/"+configID, &resp); err != nil {
				return err
			}

			if resp.ConfigVersion.Config == nil {
				return fmt.Errorf("remote config %s has no config data", configID)
			}

			configJSON, err := json.Marshal(resp.ConfigVersion.Config)
			if err != nil {
				return fmt.Errorf("marshaling remote config: %w", err)
			}
			remoteCfg, err := configascode.FromJSON(configJSON)
			if err != nil {
				return fmt.Errorf("parsing remote config: %w", err)
			}

			result, err := configascode.Diff(&cfgFile.Config, remoteCfg)
			if err != nil {
				return err
			}

			if a.Format == "json" {
				return output.Write(a.Out, "json", result, nil)
			}

			if err := configascode.WriteDiff(a.Out, result); err != nil {
				return err
			}

			_, err = fmt.Fprintf(a.Out, "\nSummary: %s\n", configascode.SummaryLine(result))
			return err
		},
	}

	cmd.Flags().StringVar(&configID, "id", "", "Remote config version ID to compare against (required)")
	return cmd
}
