package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/sable-inc/anvil/internal/output"
)

// Agent mirrors the sable-api Agent response shape.
type Agent struct {
	ID        int    `json:"id" yaml:"id"`
	PublicID  string `json:"publicId" yaml:"publicId"`
	OrgID     int    `json:"orgId" yaml:"orgId"`
	Name      string `json:"name" yaml:"name"`
	Slug      string `json:"slug" yaml:"slug"`
	Status    string `json:"status" yaml:"status"`
	CreatedAt string `json:"createdAt" yaml:"createdAt"`
	UpdatedAt string `json:"updatedAt" yaml:"updatedAt"`
}

func newAgentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "agent",
		Aliases: []string{"agents"},
		Short:   "Manage agents",
	}

	cmd.AddCommand(newAgentListCmd())
	cmd.AddCommand(newAgentGetCmd())
	cmd.AddCommand(newAgentCreateCmd())
	cmd.AddCommand(newAgentUpdateCmd())
	cmd.AddCommand(newAgentDeleteCmd())
	cmd.AddCommand(newAgentPullConfigCmd())
	return cmd
}

func newAgentListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all agents",
		RunE: func(cmd *cobra.Command, _ []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}

			var resp struct {
				Agents []Agent `json:"agents"`
			}
			path := "/agents"
			if a.OrgID != "" {
				path += "?orgId=" + a.OrgID
			}
			if err := client.Get(cmd.Context(), path, &resp); err != nil {
				return err
			}

			t := output.NewTable("ID", "Public ID", "Name", "Slug", "Status")
			for _, ag := range resp.Agents {
				t.AddRow(strconv.Itoa(ag.ID), ag.PublicID, ag.Name, ag.Slug, ag.Status)
			}
			return output.Write(a.Out, a.Format, resp.Agents, t)
		},
	}
}

func newAgentGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:               "get <id|slug|publicId>",
		Short:             "Get agent details",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeAgents,
		RunE: func(cmd *cobra.Command, args []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}

			var resp struct {
				Agent Agent `json:"agent"`
			}
			if err := client.Get(cmd.Context(), "/agents/"+args[0], &resp); err != nil {
				return err
			}

			ag := resp.Agent
			t := output.NewTable("Field", "Value")
			t.AddRow("ID", strconv.Itoa(ag.ID))
			t.AddRow("Public ID", ag.PublicID)
			t.AddRow("Name", ag.Name)
			t.AddRow("Slug", ag.Slug)
			t.AddRow("Status", ag.Status)
			t.AddRow("Org ID", strconv.Itoa(ag.OrgID))
			t.AddRow("Created", ag.CreatedAt)
			t.AddRow("Updated", ag.UpdatedAt)
			return output.Write(a.Out, a.Format, ag, t)
		},
	}
}

func newAgentCreateCmd() *cobra.Command {
	var name, slug, status string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an agent",
		RunE: func(cmd *cobra.Command, _ []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}

			body := map[string]any{"name": name}
			if slug != "" {
				body["slug"] = slug
			}
			if status != "" {
				body["status"] = status
			}

			var resp struct {
				Agent Agent `json:"agent"`
			}
			if err := client.Post(cmd.Context(), "/agents", body, &resp); err != nil {
				return err
			}

			ag := resp.Agent
			t := output.NewTable("Field", "Value")
			t.AddRow("ID", strconv.Itoa(ag.ID))
			t.AddRow("Name", ag.Name)
			t.AddRow("Slug", ag.Slug)
			t.AddRow("Status", ag.Status)
			return output.Write(a.Out, a.Format, ag, t)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Agent name (required)")
	cmd.Flags().StringVar(&slug, "slug", "", "Agent slug")
	cmd.Flags().StringVar(&status, "status", "", "Agent status (active|inactive|archived)")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newAgentUpdateCmd() *cobra.Command {
	var name, slug, status string

	cmd := &cobra.Command{
		Use:               "update <id>",
		Short:             "Update an agent",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeAgents,
		RunE: func(cmd *cobra.Command, args []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}

			body := map[string]any{}
			if name != "" {
				body["name"] = name
			}
			if slug != "" {
				body["slug"] = slug
			}
			if status != "" {
				body["status"] = status
			}
			if len(body) == 0 {
				return fmt.Errorf("at least one of --name, --slug, or --status is required")
			}

			var resp struct {
				Agent Agent `json:"agent"`
			}
			if err := client.Put(cmd.Context(), "/agents/"+args[0], body, &resp); err != nil {
				return err
			}

			ag := resp.Agent
			_, err = fmt.Fprintf(a.Out, "Agent %q updated (ID: %d)\n", ag.Name, ag.ID)
			return err
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Agent name")
	cmd.Flags().StringVar(&slug, "slug", "", "Agent slug")
	cmd.Flags().StringVar(&status, "status", "", "Agent status")
	return cmd
}

func newAgentPullConfigCmd() *cobra.Command {
	var outFile string

	cmd := &cobra.Command{
		Use:               "pull-config <agentId|publicId|slug>",
		Short:             "Download the published config for an agent",
		Long:              "Fetches the published agent config by ID, public ID, or slug.\nUse -o to write to a file for local development.",
		Example:           "  anvil agent pull-config agt_eNVj4PXnHSTSLyJYdojjw -o agent.json",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeAgents,
		RunE: func(cmd *cobra.Command, args []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireClient()
			if err != nil {
				return err
			}

			var resp struct {
				Config json.RawMessage `json:"config"`
			}
			if err := client.Get(cmd.Context(), "/agents/"+args[0]+"/published-config", &resp); err != nil {
				return err
			}

			if resp.Config == nil {
				return fmt.Errorf("no published config found for agent %s", args[0])
			}

			// Pretty-print the config JSON.
			var pretty json.RawMessage
			if err := json.Unmarshal(resp.Config, &pretty); err != nil {
				return fmt.Errorf("parsing config: %w", err)
			}
			data, err := json.MarshalIndent(pretty, "", "  ")
			if err != nil {
				return fmt.Errorf("formatting config: %w", err)
			}

			if outFile != "" {
				if err := os.WriteFile(outFile, append(data, '\n'), 0o644); err != nil { //nolint:gosec // user writes their own config
					return fmt.Errorf("writing file: %w", err)
				}
				_, err = fmt.Fprintf(a.Out, "Config written to %s\n", outFile)
				return err
			}

			_, err = fmt.Fprintf(a.Out, "%s\n", data)
			return err
		},
	}

	cmd.Flags().StringVarP(&outFile, "output", "o", "", "Write config to file (default: stdout)")
	return cmd
}

func newAgentDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:               "delete <id>",
		Short:             "Delete an agent (admin only)",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeAgents,
		RunE: func(cmd *cobra.Command, args []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}

			if err := client.Delete(cmd.Context(), "/agents/"+args[0], nil); err != nil {
				return err
			}
			_, err = fmt.Fprintf(a.Out, "Agent %s deleted.\n", args[0])
			return err
		},
	}
}
