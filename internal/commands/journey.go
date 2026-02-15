package commands

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/sable-inc/anvil/internal/output"
)

// Journey mirrors the sable-api Journey response shape.
type Journey struct {
	ID          int    `json:"id" yaml:"id"`
	PublicID    string `json:"publicId" yaml:"publicId"`
	OrgID       int    `json:"orgId" yaml:"orgId"`
	AgentID     *int   `json:"agentId" yaml:"agentId"`
	Name        string `json:"name" yaml:"name"`
	Slug        string `json:"slug" yaml:"slug"`
	Description string `json:"description" yaml:"description"`
	Version     int    `json:"version" yaml:"version"`
	CreatedAt   string `json:"createdAt" yaml:"createdAt"`
	UpdatedAt   string `json:"updatedAt" yaml:"updatedAt"`
}

// JourneyDetail includes moments and transitions.
type JourneyDetail struct {
	Journey
	Moments     []any `json:"moments" yaml:"moments"`
	Transitions []any `json:"transitions" yaml:"transitions"`
}

func newJourneyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "journey",
		Aliases: []string{"journeys"},
		Short:   "Manage journeys",
	}

	cmd.AddCommand(newJourneyListCmd())
	cmd.AddCommand(newJourneyGetCmd())
	cmd.AddCommand(newJourneyCreateCmd())
	cmd.AddCommand(newJourneyUpdateCmd())
	cmd.AddCommand(newJourneyDeleteCmd())
	return cmd
}

func newJourneyListCmd() *cobra.Command {
	var agentID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List journeys",
		RunE: func(cmd *cobra.Command, _ []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}

			path := "/journeys"
			sep := "?"
			if a.OrgID != "" {
				path += sep + "orgId=" + a.OrgID
				sep = "&"
			}
			if agentID != "" {
				path += sep + "agentId=" + agentID
			}

			var resp struct {
				Journeys []Journey `json:"journeys"`
			}
			if err := client.Get(cmd.Context(), path, &resp); err != nil {
				return err
			}

			t := output.NewTable("ID", "Public ID", "Name", "Slug", "Version", "Agent ID")
			for _, j := range resp.Journeys {
				aid := "-"
				if j.AgentID != nil {
					aid = strconv.Itoa(*j.AgentID)
				}
				t.AddRow(strconv.Itoa(j.ID), j.PublicID, j.Name, j.Slug, strconv.Itoa(j.Version), aid)
			}
			return output.Write(a.Out, a.Format, resp.Journeys, t)
		},
	}

	cmd.Flags().StringVar(&agentID, "agent-id", "", "Filter by agent ID")
	return cmd
}

func newJourneyGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:               "get <id|slug|publicId>",
		Short:             "Get journey details (with moments and transitions)",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeJourneys,
		RunE: func(cmd *cobra.Command, args []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}

			var resp struct {
				Journey JourneyDetail `json:"journey"`
			}
			if err := client.Get(cmd.Context(), "/journeys/"+args[0], &resp); err != nil {
				return err
			}

			j := resp.Journey
			t := output.NewTable("Field", "Value")
			t.AddRow("ID", strconv.Itoa(j.ID))
			t.AddRow("Public ID", j.PublicID)
			t.AddRow("Name", j.Name)
			t.AddRow("Slug", j.Slug)
			t.AddRow("Description", j.Description)
			t.AddRow("Version", strconv.Itoa(j.Version))
			t.AddRow("Moments", strconv.Itoa(len(j.Moments)))
			t.AddRow("Transitions", strconv.Itoa(len(j.Transitions)))
			t.AddRow("Created", j.CreatedAt)
			t.AddRow("Updated", j.UpdatedAt)
			return output.Write(a.Out, a.Format, j, t)
		},
	}
}

func newJourneyCreateCmd() *cobra.Command {
	var name, slug, description string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a journey",
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
			if description != "" {
				body["description"] = description
			}

			var resp struct {
				Journey Journey `json:"journey"`
			}
			if err := client.Post(cmd.Context(), "/journeys", body, &resp); err != nil {
				return err
			}

			j := resp.Journey
			t := output.NewTable("Field", "Value")
			t.AddRow("ID", strconv.Itoa(j.ID))
			t.AddRow("Name", j.Name)
			t.AddRow("Slug", j.Slug)
			return output.Write(a.Out, a.Format, j, t)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Journey name (required)")
	cmd.Flags().StringVar(&slug, "slug", "", "Journey slug")
	cmd.Flags().StringVar(&description, "description", "", "Journey description")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newJourneyUpdateCmd() *cobra.Command {
	var name, slug, description string

	cmd := &cobra.Command{
		Use:               "update <id>",
		Short:             "Update a journey",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeJourneys,
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
			if description != "" {
				body["description"] = description
			}
			if len(body) == 0 {
				return fmt.Errorf("at least one of --name, --slug, or --description is required")
			}

			var resp struct {
				Journey Journey `json:"journey"`
			}
			if err := client.Put(cmd.Context(), "/journeys/"+args[0], body, &resp); err != nil {
				return err
			}

			j := resp.Journey
			_, err = fmt.Fprintf(a.Out, "Journey %q updated (ID: %d)\n", j.Name, j.ID)
			return err
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Journey name")
	cmd.Flags().StringVar(&slug, "slug", "", "Journey slug")
	cmd.Flags().StringVar(&description, "description", "", "Journey description")
	return cmd
}

func newJourneyDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:               "delete <id>",
		Short:             "Delete a journey",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeJourneys,
		RunE: func(cmd *cobra.Command, args []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}

			if err := client.Delete(cmd.Context(), "/journeys/"+args[0], nil); err != nil {
				return err
			}
			_, err = fmt.Fprintf(a.Out, "Journey %s deleted.\n", args[0])
			return err
		},
	}
}
