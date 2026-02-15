package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/sable-inc/anvil/internal/version"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print Anvil version information",
		RunE: func(cmd *cobra.Command, _ []string) error {
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]string{
					"version": version.Version,
					"commit":  version.Commit,
					"date":    version.Date,
				})
			}
			_, err := fmt.Fprintln(cmd.OutOrStdout(), version.Info())
			return err
		},
	}
}
