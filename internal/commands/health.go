package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/sable-inc/anvil/internal/output"
)

func newHealthCmd() *cobra.Command {
	var withDB bool

	cmd := &cobra.Command{
		Use:   "health",
		Short: "Check API health",
		RunE: func(cmd *cobra.Command, _ []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireClient()
			if err != nil {
				return err
			}

			type healthResp struct {
				Status string `json:"status"`
			}

			var api healthResp
			if err := client.Get(cmd.Context(), "/health", &api); err != nil {
				return fmt.Errorf("API health check failed: %w", err)
			}

			var db healthResp
			if withDB {
				if err := client.Get(cmd.Context(), "/db/health", &db); err != nil {
					return fmt.Errorf("database health check failed: %w", err)
				}
			}

			t := output.NewTable("Service", "Status")
			t.AddRow("api", api.Status)
			if withDB {
				t.AddRow("database", db.Status)
			}

			data := map[string]string{"api": api.Status}
			if withDB {
				data["database"] = db.Status
			}
			return output.Write(a.Out, a.Format, data, t)
		},
	}

	cmd.Flags().BoolVar(&withDB, "db", false, "Also check database health")
	return cmd
}
