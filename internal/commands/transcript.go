package commands

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/sable-inc/anvil/internal/output"
)

// TranscriptSession mirrors the sable-api transcript session shape.
type TranscriptSession struct {
	ID             string `json:"id" yaml:"id"`
	SessionID      string `json:"sessionId" yaml:"sessionId"`
	ModuleID       string `json:"moduleId" yaml:"moduleId"`
	ModuleName     string `json:"moduleName" yaml:"moduleName"`
	AgentID        string `json:"agentId" yaml:"agentId"`
	UserID         string `json:"userId" yaml:"userId"`
	PartnerName    string `json:"partnerName" yaml:"partnerName"`
	PartnerCompany string `json:"partnerCompany" yaml:"partnerCompany"`
	CreatedAt      string `json:"createdAt" yaml:"createdAt"`
	Messages       []TranscriptMessage `json:"messages" yaml:"messages"`
}

// TranscriptMessage is a single message in a transcript.
type TranscriptMessage struct {
	Type      string `json:"type" yaml:"type"`
	Text      string `json:"text" yaml:"text"`
	Timestamp string `json:"timestamp" yaml:"timestamp"`
}

func newTranscriptCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "transcript",
		Aliases: []string{"transcripts"},
		Short:   "View conversation transcripts",
	}

	cmd.AddCommand(newTranscriptListCmd())
	cmd.AddCommand(newTranscriptViewCmd())
	return cmd
}

func newTranscriptListCmd() *cobra.Command {
	var limit string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List transcript sessions",
		RunE: func(cmd *cobra.Command, _ []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}

			path := "/transcripts"
			sep := "?"
			if a.OrgID != "" {
				path += sep + "orgId=" + a.OrgID
				sep = "&"
			}
			if limit != "" {
				path += sep + "limit=" + limit
			}

			var resp struct {
				Transcripts []TranscriptSession `json:"transcripts"`
			}
			if err := client.Get(cmd.Context(), path, &resp); err != nil {
				return err
			}

			t := output.NewTable("ID", "Module", "Partner", "Company", "Messages", "Created")
			for _, ts := range resp.Transcripts {
				t.AddRow(
					ts.ID,
					ts.ModuleName,
					ts.PartnerName,
					ts.PartnerCompany,
					strconv.Itoa(len(ts.Messages)),
					ts.CreatedAt,
				)
			}
			return output.Write(a.Out, a.Format, resp.Transcripts, t)
		},
	}

	cmd.Flags().StringVar(&limit, "limit", "", "Max number of transcripts to return")
	return cmd
}

func newTranscriptViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "view <id>",
		Short: "View a transcript session with messages",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}

			// Fetch the transcript list and find the matching session.
			path := "/transcripts"
			if a.OrgID != "" {
				path += "?orgId=" + a.OrgID
			}

			var resp struct {
				Transcripts []TranscriptSession `json:"transcripts"`
			}
			if err := client.Get(cmd.Context(), path, &resp); err != nil {
				return err
			}

			var session *TranscriptSession
			for i := range resp.Transcripts {
				if resp.Transcripts[i].ID == args[0] || resp.Transcripts[i].SessionID == args[0] {
					session = &resp.Transcripts[i]
					break
				}
			}
			if session == nil {
				return fmt.Errorf("transcript session %q not found", args[0])
			}

			// For JSON/YAML, output the full session.
			if a.Format == "json" || a.Format == "yaml" {
				f := output.New(a.Format)
				return f.Format(a.Out, session)
			}

			// Table format: show metadata then messages.
			if _, err := fmt.Fprintf(a.Out, "Session: %s\nModule:  %s\nPartner: %s (%s)\n\n",
				session.SessionID, session.ModuleName,
				session.PartnerName, session.PartnerCompany,
			); err != nil {
				return err
			}

			t := output.NewTable("Time", "Sender", "Message")
			for _, m := range session.Messages {
				text := m.Text
				if len(text) > 120 {
					text = text[:117] + "..."
				}
				t.AddRow(m.Timestamp, m.Type, text)
			}
			return t.Render(a.Out)
		},
	}
}
