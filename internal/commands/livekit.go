package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/sable-inc/anvil/internal/output"
)

// LiveKit response types.

// Room represents a LiveKit room.
type Room struct {
	Name            string `json:"name" yaml:"name"`
	SID             string `json:"sid" yaml:"sid"`
	NumParticipants int    `json:"numParticipants" yaml:"numParticipants"`
	CreationTime    int64  `json:"creationTime" yaml:"creationTime"`
	Metadata        string `json:"metadata" yaml:"metadata"`
}

// Participant represents a LiveKit room participant.
type Participant struct {
	Identity string `json:"identity" yaml:"identity"`
	SID      string `json:"sid" yaml:"sid"`
	State    int    `json:"state" yaml:"state"`
	JoinedAt int64  `json:"joinedAt" yaml:"joinedAt"`
	Name     string `json:"name" yaml:"name"`
	Metadata string `json:"metadata" yaml:"metadata"`
}

// AgentSummary is a brief agent entry returned from the agents list endpoint.
type AgentSummary struct {
	ID     string  `json:"id" yaml:"id"`
	Name   string  `json:"name" yaml:"name"`
	Status *string `json:"status" yaml:"status"`
}

// AgentStatus describes the current state of a LiveKit agent.
type AgentStatus struct {
	Status         string  `json:"status" yaml:"status"`
	Replicas       *int    `json:"replicas" yaml:"replicas"`
	CurrentVersion *string `json:"currentVersion" yaml:"currentVersion"`
	Uptime         *string `json:"uptime" yaml:"uptime"`
	Image          *string `json:"image" yaml:"image"`
	Raw            string  `json:"raw" yaml:"raw"`
}

// AgentVersion describes a versioned release of a LiveKit agent.
type AgentVersion struct {
	VersionID string  `json:"versionId" yaml:"versionId"`
	CreatedAt string  `json:"createdAt" yaml:"createdAt"`
	Status    string  `json:"status" yaml:"status"`
	Image     *string `json:"image" yaml:"image"`
}

// livekitHeaders holds optional credential overrides from flags.
type livekitHeaders struct {
	url       string
	apiKey    string
	apiSecret string
}

func (h livekitHeaders) toMap() map[string]string {
	m := map[string]string{}
	if h.url != "" {
		m["x-livekit-url"] = h.url
	}
	if h.apiKey != "" {
		m["x-livekit-api-key"] = h.apiKey
	}
	if h.apiSecret != "" {
		m["x-livekit-api-secret"] = h.apiSecret
	}
	return m
}

func newLivekitCmd() *cobra.Command {
	var lkh livekitHeaders

	cmd := &cobra.Command{
		Use:     "livekit",
		Aliases: []string{"lk"},
		Short:   "LiveKit infrastructure operations",
		Long:    "Manage LiveKit sessions, agents, and secrets.",
	}

	cmd.PersistentFlags().StringVar(&lkh.url, "livekit-url", "", "Override LiveKit URL")
	cmd.PersistentFlags().StringVar(&lkh.apiKey, "livekit-api-key", "", "Override LiveKit API key")
	cmd.PersistentFlags().StringVar(&lkh.apiSecret, "livekit-api-secret", "", "Override LiveKit API secret")

	cmd.AddCommand(newLivekitSessionsCmd(&lkh))
	cmd.AddCommand(newLivekitAgentCmd(&lkh))
	return cmd
}

// ─── Sessions ────────────────────────────────────────────────────────────────

func newLivekitSessionsCmd(lkh *livekitHeaders) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sessions",
		Aliases: []string{"session", "rooms"},
		Short:   "Manage LiveKit sessions (rooms)",
	}

	cmd.AddCommand(newLivekitSessionsListCmd(lkh))
	cmd.AddCommand(newLivekitSessionsGetCmd(lkh))
	cmd.AddCommand(newLivekitSessionsCloseCmd(lkh))
	cmd.AddCommand(newLivekitSessionsRemoveParticipantCmd(lkh))
	cmd.AddCommand(newLivekitSessionsMuteCmd(lkh))
	return cmd
}

func newLivekitSessionsListCmd(lkh *livekitHeaders) *cobra.Command {
	var (
		agentID     string
		environment string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List active sessions",
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

			path := "/organizations/" + publicID + "/livekit/sessions"
			q := buildQuery(agentID, environment)
			if q != "" {
				path += "?" + q
			}

			var resp struct {
				Rooms []Room `json:"rooms"`
			}
			headers := lkh.toMap()
			if err := client.GetWithHeaders(cmd.Context(), path, headers, &resp); err != nil {
				return err
			}

			t := output.NewTable("Name", "SID", "Participants", "Created")
			for _, r := range resp.Rooms {
				t.AddRow(r.Name, r.SID, strconv.Itoa(r.NumParticipants), strconv.FormatInt(r.CreationTime, 10))
			}
			return output.Write(a.Out, a.Format, resp.Rooms, t)
		},
	}

	cmd.Flags().StringVar(&agentID, "agent-id", "", "Filter by agent ID")
	cmd.Flags().StringVar(&environment, "environment", "", "Environment (production|test)")
	return cmd
}

func newLivekitSessionsGetCmd(lkh *livekitHeaders) *cobra.Command {
	return &cobra.Command{
		Use:   "get <roomName>",
		Short: "Get session details with participants",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}
			publicID, err := a.RequirePublicID(cmd.Context())
			if err != nil {
				return err
			}

			path := "/organizations/" + publicID + "/livekit/sessions/" + args[0]
			var resp struct {
				Room         Room          `json:"room"`
				Participants []Participant `json:"participants"`
			}
			headers := lkh.toMap()
			if err := client.GetWithHeaders(cmd.Context(), path, headers, &resp); err != nil {
				return err
			}

			t := output.NewTable("Field", "Value")
			t.AddRow("Room Name", resp.Room.Name)
			t.AddRow("SID", resp.Room.SID)
			t.AddRow("Participants", strconv.Itoa(resp.Room.NumParticipants))
			for i, p := range resp.Participants {
				prefix := fmt.Sprintf("Participant %d", i+1)
				t.AddRow(prefix+" Identity", p.Identity)
				t.AddRow(prefix+" Name", p.Name)
				t.AddRow(prefix+" State", strconv.Itoa(p.State))
			}
			return output.Write(a.Out, a.Format, resp, t)
		},
	}
}

func newLivekitSessionsCloseCmd(lkh *livekitHeaders) *cobra.Command {
	return &cobra.Command{
		Use:   "close <roomName>",
		Short: "Close a session (delete room)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}
			publicID, err := a.RequirePublicID(cmd.Context())
			if err != nil {
				return err
			}

			path := "/organizations/" + publicID + "/livekit/sessions/" + args[0]
			var resp struct {
				Message string `json:"message"`
			}
			headers := lkh.toMap()
			if err := client.DeleteWithHeaders(cmd.Context(), path, headers, &resp); err != nil {
				return err
			}
			_, err = fmt.Fprintln(a.Out, resp.Message)
			return err
		},
	}
}

func newLivekitSessionsRemoveParticipantCmd(lkh *livekitHeaders) *cobra.Command {
	return &cobra.Command{
		Use:   "remove-participant <roomName> <identity>",
		Short: "Remove a participant from a session",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}
			publicID, err := a.RequirePublicID(cmd.Context())
			if err != nil {
				return err
			}

			path := "/organizations/" + publicID + "/livekit/sessions/" + args[0] + "/participants/" + args[1]
			var resp struct {
				Message string `json:"message"`
			}
			headers := lkh.toMap()
			if err := client.DeleteWithHeaders(cmd.Context(), path, headers, &resp); err != nil {
				return err
			}
			_, err = fmt.Fprintln(a.Out, resp.Message)
			return err
		},
	}
}

func newLivekitSessionsMuteCmd(lkh *livekitHeaders) *cobra.Command {
	var (
		trackSID string
		muted    bool
	)

	cmd := &cobra.Command{
		Use:   "mute <roomName> <identity>",
		Short: "Mute or unmute a participant's track",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}
			publicID, err := a.RequirePublicID(cmd.Context())
			if err != nil {
				return err
			}

			path := "/organizations/" + publicID + "/livekit/sessions/" + args[0] + "/participants/" + args[1] + "/mute"
			body := map[string]any{
				"trackSid": trackSID,
				"muted":    muted,
			}
			var resp struct {
				Message string `json:"message"`
			}
			headers := lkh.toMap()
			if err := client.PostWithHeaders(cmd.Context(), path, body, headers, &resp); err != nil {
				return err
			}
			_, err = fmt.Fprintln(a.Out, resp.Message)
			return err
		},
	}

	cmd.Flags().StringVar(&trackSID, "track-sid", "", "Track SID to mute (required)")
	cmd.Flags().BoolVar(&muted, "muted", true, "Mute (true) or unmute (false)")
	_ = cmd.MarkFlagRequired("track-sid")
	return cmd
}

// ─── Agent ───────────────────────────────────────────────────────────────────

func newLivekitAgentCmd(lkh *livekitHeaders) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "agent",
		Aliases: []string{"agents"},
		Short:   "Manage LiveKit agents",
	}

	cmd.AddCommand(newLivekitAgentListCmd(lkh))
	cmd.AddCommand(newLivekitAgentStatusCmd(lkh))
	cmd.AddCommand(newLivekitAgentVersionsCmd(lkh))
	cmd.AddCommand(newLivekitAgentLogsCmd(lkh))
	cmd.AddCommand(newLivekitAgentSecretsCmd(lkh))
	cmd.AddCommand(newLivekitAgentRestartCmd(lkh))
	cmd.AddCommand(newLivekitAgentDeleteCmd(lkh))
	return cmd
}

func newLivekitAgentListCmd(lkh *livekitHeaders) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List LiveKit agents",
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

			path := "/organizations/" + publicID + "/livekit/agents"
			var resp struct {
				Agents []AgentSummary `json:"agents"`
			}
			headers := lkh.toMap()
			if err := client.GetWithHeaders(cmd.Context(), path, headers, &resp); err != nil {
				return err
			}

			t := output.NewTable("ID", "Name", "Status")
			for _, ag := range resp.Agents {
				t.AddRow(ag.ID, ag.Name, ptrStr(ag.Status))
			}
			return output.Write(a.Out, a.Format, resp.Agents, t)
		},
	}
}

func newLivekitAgentStatusCmd(lkh *livekitHeaders) *cobra.Command {
	var (
		agentID     string
		environment string
	)

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Get agent status",
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

			path := "/organizations/" + publicID + "/livekit/agent/status"
			q := buildQuery(agentID, environment)
			if q != "" {
				path += "?" + q
			}

			var resp struct {
				Status AgentStatus `json:"status"`
			}
			headers := lkh.toMap()
			if err := client.GetWithHeaders(cmd.Context(), path, headers, &resp); err != nil {
				return err
			}

			s := resp.Status
			t := output.NewTable("Field", "Value")
			t.AddRow("Status", s.Status)
			if s.Replicas != nil {
				t.AddRow("Replicas", strconv.Itoa(*s.Replicas))
			}
			if s.CurrentVersion != nil {
				t.AddRow("Version", *s.CurrentVersion)
			}
			if s.Uptime != nil {
				t.AddRow("Uptime", *s.Uptime)
			}
			if s.Image != nil {
				t.AddRow("Image", *s.Image)
			}
			return output.Write(a.Out, a.Format, s, t)
		},
	}

	cmd.Flags().StringVar(&agentID, "agent-id", "", "Agent ID")
	cmd.Flags().StringVar(&environment, "environment", "", "Environment (production|test)")
	return cmd
}

func newLivekitAgentVersionsCmd(lkh *livekitHeaders) *cobra.Command {
	var (
		agentID     string
		environment string
	)

	cmd := &cobra.Command{
		Use:   "versions",
		Short: "List agent versions",
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

			path := "/organizations/" + publicID + "/livekit/agent/versions"
			q := buildQuery(agentID, environment)
			if q != "" {
				path += "?" + q
			}

			var resp struct {
				Versions []AgentVersion `json:"versions"`
				Raw      string         `json:"raw"`
			}
			headers := lkh.toMap()
			if err := client.GetWithHeaders(cmd.Context(), path, headers, &resp); err != nil {
				return err
			}

			t := output.NewTable("Version", "Status", "Created", "Image")
			for _, v := range resp.Versions {
				t.AddRow(v.VersionID, v.Status, v.CreatedAt, ptrStr(v.Image))
			}
			return output.Write(a.Out, a.Format, resp.Versions, t)
		},
	}

	cmd.Flags().StringVar(&agentID, "agent-id", "", "Agent ID")
	cmd.Flags().StringVar(&environment, "environment", "", "Environment (production|test)")
	return cmd
}

func newLivekitAgentLogsCmd(lkh *livekitHeaders) *cobra.Command {
	var (
		agentID       string
		environment   string
		logType       string
		captureTimeMs int
	)

	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Get agent logs",
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

			path := "/organizations/" + publicID + "/livekit/agent/logs"
			params := []string{}
			if agentID != "" {
				params = append(params, "agentId="+agentID)
			}
			if environment != "" {
				params = append(params, "environment="+environment)
			}
			if logType != "" {
				params = append(params, "logType="+logType)
			}
			if captureTimeMs > 0 {
				params = append(params, "captureTimeMs="+strconv.Itoa(captureTimeMs))
			}
			if len(params) > 0 {
				path += "?" + strings.Join(params, "&")
			}

			var resp struct {
				Logs    string `json:"logs"`
				LogType string `json:"logType"`
			}
			headers := lkh.toMap()
			if err := client.GetWithHeaders(cmd.Context(), path, headers, &resp); err != nil {
				return err
			}

			// For logs, raw text output is most useful.
			if a.Format == "table" {
				_, err = fmt.Fprint(a.Out, resp.Logs)
				return err
			}
			return output.Write(a.Out, a.Format, resp, nil)
		},
	}

	cmd.Flags().StringVar(&agentID, "agent-id", "", "Agent ID")
	cmd.Flags().StringVar(&environment, "environment", "", "Environment (production|test)")
	cmd.Flags().StringVar(&logType, "log-type", "", "Log type: deploy or build")
	cmd.Flags().IntVar(&captureTimeMs, "capture-time-ms", 0, "Capture time in milliseconds (1000-30000)")
	return cmd
}

func newLivekitAgentSecretsCmd(lkh *livekitHeaders) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secrets",
		Short: "Manage agent secrets",
	}

	cmd.AddCommand(newLivekitAgentSecretsListCmd(lkh))
	cmd.AddCommand(newLivekitAgentSecretsSetCmd(lkh))
	cmd.AddCommand(newLivekitAgentSecretsDeleteCmd(lkh))
	return cmd
}

func newLivekitAgentSecretsListCmd(lkh *livekitHeaders) *cobra.Command {
	var (
		agentID     string
		environment string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List agent secret names",
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

			path := "/organizations/" + publicID + "/livekit/agent/secrets"
			q := buildQuery(agentID, environment)
			if q != "" {
				path += "?" + q
			}

			var resp struct {
				Secrets []string `json:"secrets"`
			}
			headers := lkh.toMap()
			if err := client.GetWithHeaders(cmd.Context(), path, headers, &resp); err != nil {
				return err
			}

			t := output.NewTable("Secret Name")
			for _, s := range resp.Secrets {
				t.AddRow(s)
			}
			return output.Write(a.Out, a.Format, resp.Secrets, t)
		},
	}

	cmd.Flags().StringVar(&agentID, "agent-id", "", "Agent ID")
	cmd.Flags().StringVar(&environment, "environment", "", "Environment (production|test)")
	return cmd
}

func newLivekitAgentSecretsSetCmd(lkh *livekitHeaders) *cobra.Command {
	var (
		agentID     string
		environment string
		secrets     []string
	)

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set agent secrets (KEY=VALUE pairs)",
		Long:  "Update agent secrets. Use --secret KEY=VALUE (repeatable).",
		Example: `  anvil livekit agent secrets set --org org_xxx --secret OPENAI_API_KEY=sk-xxx
  anvil livekit agent secrets set --org org_xxx --secret KEY1=val1 --secret KEY2=val2`,
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

			secretMap := map[string]string{}
			for _, s := range secrets {
				k, v, ok := strings.Cut(s, "=")
				if !ok {
					return fmt.Errorf("invalid secret format %q — use KEY=VALUE", s)
				}
				secretMap[k] = v
			}
			if len(secretMap) == 0 {
				return fmt.Errorf("at least one --secret is required")
			}

			path := "/organizations/" + publicID + "/livekit/agent/secrets"
			q := buildQuery(agentID, environment)
			if q != "" {
				path += "?" + q
			}

			body := map[string]any{"secrets": secretMap}
			var resp struct {
				Message string `json:"message"`
			}
			headers := lkh.toMap()
			if err := client.PostWithHeaders(cmd.Context(), path, body, headers, &resp); err != nil {
				return err
			}
			_, err = fmt.Fprintln(a.Out, resp.Message)
			return err
		},
	}

	cmd.Flags().StringVar(&agentID, "agent-id", "", "Agent ID")
	cmd.Flags().StringVar(&environment, "environment", "", "Environment (production|test)")
	cmd.Flags().StringArrayVar(&secrets, "secret", nil, "Secret in KEY=VALUE format (repeatable)")
	return cmd
}

func newLivekitAgentSecretsDeleteCmd(lkh *livekitHeaders) *cobra.Command {
	var (
		agentID     string
		environment string
	)

	cmd := &cobra.Command{
		Use:   "delete <secretName>",
		Short: "Delete an agent secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}
			publicID, err := a.RequirePublicID(cmd.Context())
			if err != nil {
				return err
			}

			path := "/organizations/" + publicID + "/livekit/agent/secrets/" + args[0]
			q := buildQuery(agentID, environment)
			if q != "" {
				path += "?" + q
			}

			var resp struct {
				Message string `json:"message"`
			}
			headers := lkh.toMap()
			if err := client.DeleteWithHeaders(cmd.Context(), path, headers, &resp); err != nil {
				return err
			}
			_, err = fmt.Fprintln(a.Out, resp.Message)
			return err
		},
	}

	cmd.Flags().StringVar(&agentID, "agent-id", "", "Agent ID")
	cmd.Flags().StringVar(&environment, "environment", "", "Environment (production|test)")
	return cmd
}

func newLivekitAgentRestartCmd(lkh *livekitHeaders) *cobra.Command {
	var (
		agentID     string
		environment string
	)

	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart an agent",
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

			path := "/organizations/" + publicID + "/livekit/agent/restart"
			q := buildQuery(agentID, environment)
			if q != "" {
				path += "?" + q
			}

			var resp struct {
				Message string `json:"message"`
			}
			headers := lkh.toMap()
			if err := client.PostWithHeaders(cmd.Context(), path, nil, headers, &resp); err != nil {
				return err
			}
			_, err = fmt.Fprintln(a.Out, resp.Message)
			return err
		},
	}

	cmd.Flags().StringVar(&agentID, "agent-id", "", "Agent ID")
	cmd.Flags().StringVar(&environment, "environment", "", "Environment (production|test)")
	return cmd
}

func newLivekitAgentDeleteCmd(lkh *livekitHeaders) *cobra.Command {
	var (
		agentID     string
		environment string
	)

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an agent",
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

			path := "/organizations/" + publicID + "/livekit/agent"
			q := buildQuery(agentID, environment)
			if q != "" {
				path += "?" + q
			}

			var resp struct {
				Message string `json:"message"`
			}
			headers := lkh.toMap()
			if err := client.DeleteWithHeaders(cmd.Context(), path, headers, &resp); err != nil {
				return err
			}
			_, err = fmt.Fprintln(a.Out, resp.Message)
			return err
		},
	}

	cmd.Flags().StringVar(&agentID, "agent-id", "", "Agent ID")
	cmd.Flags().StringVar(&environment, "environment", "", "Environment (production|test)")
	return cmd
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

// buildQuery creates a query string from optional agentId and environment params.
func buildQuery(agentID, environment string) string {
	params := []string{}
	if agentID != "" {
		params = append(params, "agentId="+agentID)
	}
	if environment != "" {
		params = append(params, "environment="+environment)
	}
	return strings.Join(params, "&")
}
