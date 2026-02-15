package commands

import (
	"fmt"
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
		Use:   "get <id|slug|publicId>",
		Short: "Get agent details",
		Args:  cobra.ExactArgs(1),
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
		Use:   "update <id>",
		Short: "Update an agent",
		Args:  cobra.ExactArgs(1),
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

func newAgentDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an agent (admin only)",
		Args:  cobra.ExactArgs(1),
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
