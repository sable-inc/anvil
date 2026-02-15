package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/sable-inc/anvil/internal/output"
)

// Organization mirrors the sable-api Organization response shape.
type Organization struct {
	ID        int      `json:"id" yaml:"id"`
	PublicID  string   `json:"publicId" yaml:"publicId"`
	Name      string   `json:"name" yaml:"name"`
	Domains   []string `json:"domains" yaml:"domains"`
	CreatedAt string   `json:"createdAt" yaml:"createdAt"`
	UpdatedAt string   `json:"updatedAt" yaml:"updatedAt"`
}

func newOrgCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "org",
		Aliases: []string{"orgs", "organization", "organizations"},
		Short:   "Manage organizations",
	}

	cmd.AddCommand(newOrgListCmd())
	cmd.AddCommand(newOrgGetCmd())
	return cmd
}

func newOrgListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all organizations",
		RunE: func(cmd *cobra.Command, _ []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}

			var resp struct {
				Organizations []Organization `json:"organizations"`
			}
			if err := client.Get(cmd.Context(), "/organizations", &resp); err != nil {
				return err
			}

			t := output.NewTable("ID", "Public ID", "Name", "Domains")
			for _, o := range resp.Organizations {
				domains := "-"
				if len(o.Domains) > 0 {
					domains = strings.Join(o.Domains, ", ")
				}
				t.AddRow(strconv.Itoa(o.ID), o.PublicID, o.Name, domains)
			}
			return output.Write(a.Out, a.Format, resp.Organizations, t)
		},
	}
}

func newOrgGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id|publicId>",
		Short: "Get organization details",
		Long:  "Get an organization by numeric ID or public ID (org_xxx).",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}

			identifier := args[0]
			var org Organization

			if strings.HasPrefix(identifier, "org_") {
				// Public ID lookup: fetch all orgs and find the match.
				var resp struct {
					Organizations []Organization `json:"organizations"`
				}
				if err := client.Get(cmd.Context(), "/organizations", &resp); err != nil {
					return err
				}
				found := false
				for _, o := range resp.Organizations {
					if o.PublicID == identifier {
						org = o
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("organization %s not found", identifier)
				}
			} else {
				// Numeric ID lookup.
				var resp struct {
					Organization Organization `json:"organization"`
				}
				if err := client.Get(cmd.Context(), "/organizations/"+identifier, &resp); err != nil {
					return err
				}
				org = resp.Organization
			}

			t := output.NewTable("Field", "Value")
			t.AddRow("ID", strconv.Itoa(org.ID))
			t.AddRow("Public ID", org.PublicID)
			t.AddRow("Name", org.Name)
			domains := "-"
			if len(org.Domains) > 0 {
				domains = strings.Join(org.Domains, ", ")
			}
			t.AddRow("Domains", domains)
			t.AddRow("Created", org.CreatedAt)
			t.AddRow("Updated", org.UpdatedAt)
			return output.Write(a.Out, a.Format, org, t)
		},
	}
}
