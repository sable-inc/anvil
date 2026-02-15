package commands

import (
	"github.com/spf13/cobra"

	"github.com/sable-inc/anvil/internal/output"
)

func newConnectCmd() *cobra.Command {
	var configID, environment string

	cmd := &cobra.Command{
		Use:               "connect <agent-slug>",
		Short:             "Get LiveKit connection details for an agent",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeAgents,
		RunE: func(cmd *cobra.Command, args []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireClient()
			if err != nil {
				return err
			}

			path := "/connection-details?agentSlug=" + args[0]
			if configID != "" {
				path += "&configId=" + configID
			}
			if environment != "" {
				path += "&environment=" + environment
			}

			var resp struct {
				ServerURL        string `json:"serverUrl"`
				RoomName         string `json:"roomName"`
				ParticipantToken string `json:"participantToken"`
				ParticipantName  string `json:"participantName"`
			}
			if err := client.Get(cmd.Context(), path, &resp); err != nil {
				return err
			}

			t := output.NewTable("Field", "Value")
			t.AddRow("Server URL", resp.ServerURL)
			t.AddRow("Room Name", resp.RoomName)
			t.AddRow("Participant", resp.ParticipantName)
			t.AddRow("Token", maskToken(resp.ParticipantToken))
			return output.Write(a.Out, a.Format, resp, t)
		},
	}

	cmd.Flags().StringVar(&configID, "config-id", "", "Config version ID")
	cmd.Flags().StringVar(&environment, "env", "", "Environment (production|test)")
	return cmd
}
