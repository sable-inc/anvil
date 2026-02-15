package commands

import (
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/sable-inc/anvil/internal/api"
)

// completionClient returns an authenticated API client for use in tab-completion.
// Returns nil silently if auth or config isn't available — completions degrade gracefully.
func completionClient(cmd *cobra.Command) *api.Client {
	a := AppFrom(cmd)
	if a == nil {
		return nil
	}
	client, err := a.RequireAuth()
	if err != nil {
		return nil
	}
	return client
}

// completeAgents provides tab-completion for agent slugs/IDs.
func completeAgents(cmd *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	client := completionClient(cmd)
	if client == nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	a := AppFrom(cmd)
	path := "/agents"
	if a.OrgID != "" {
		path += "?orgId=" + a.OrgID
	}

	var resp struct {
		Agents []Agent `json:"agents"`
	}
	if err := client.Get(cmd.Context(), path, &resp); err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var suggestions []string
	for _, ag := range resp.Agents {
		if toComplete == "" || strings.HasPrefix(ag.Slug, toComplete) {
			suggestions = append(suggestions, ag.Slug+"\t"+ag.Name)
		}
		idStr := strconv.Itoa(ag.ID)
		if strings.HasPrefix(idStr, toComplete) {
			suggestions = append(suggestions, idStr+"\t"+ag.Name)
		}
	}
	return suggestions, cobra.ShellCompDirectiveNoFileComp
}

// completeJourneys provides tab-completion for journey slugs/IDs.
func completeJourneys(cmd *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	client := completionClient(cmd)
	if client == nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	a := AppFrom(cmd)
	path := "/journeys"
	if a.OrgID != "" {
		path += "?orgId=" + a.OrgID
	}

	var resp struct {
		Journeys []Journey `json:"journeys"`
	}
	if err := client.Get(cmd.Context(), path, &resp); err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var suggestions []string
	for _, j := range resp.Journeys {
		if toComplete == "" || strings.HasPrefix(j.Slug, toComplete) {
			suggestions = append(suggestions, j.Slug+"\t"+j.Name)
		}
		idStr := strconv.Itoa(j.ID)
		if strings.HasPrefix(idStr, toComplete) {
			suggestions = append(suggestions, idStr+"\t"+j.Name)
		}
	}
	return suggestions, cobra.ShellCompDirectiveNoFileComp
}

// completeKBItems provides tab-completion for knowledge base item IDs.
func completeKBItems(cmd *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	client := completionClient(cmd)
	if client == nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	a := AppFrom(cmd)
	path := "/knowledge-base"
	if a.OrgID != "" {
		path += "?orgId=" + a.OrgID
	}

	var resp struct {
		Items []KBItem `json:"items"`
	}
	if err := client.Get(cmd.Context(), path, &resp); err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var suggestions []string
	for _, item := range resp.Items {
		idStr := strconv.Itoa(item.ID)
		if toComplete == "" || strings.HasPrefix(idStr, toComplete) {
			suggestions = append(suggestions, idStr+"\t"+item.Name+" ("+item.Type+")")
		}
	}
	return suggestions, cobra.ShellCompDirectiveNoFileComp
}

// completeConfigs provides tab-completion for agent config version IDs.
func completeConfigs(cmd *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	client := completionClient(cmd)
	if client == nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	a := AppFrom(cmd)
	path := "/agent-configs"
	if a.OrgID != "" {
		path += "?orgId=" + a.OrgID
	}

	var resp struct {
		Configs []AgentConfigVersion `json:"configs"`
	}
	if err := client.Get(cmd.Context(), path, &resp); err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var suggestions []string
	for _, c := range resp.Configs {
		if toComplete == "" || strings.HasPrefix(c.ID, toComplete) {
			suggestions = append(suggestions, c.ID+"\t"+c.Status)
		}
	}
	return suggestions, cobra.ShellCompDirectiveNoFileComp
}

// completeFormat provides tab-completion for --format flag values.
func completeFormat(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return []string{
		"table\tFormatted table output (default)",
		"json\tJSON output",
		"yaml\tYAML output",
	}, cobra.ShellCompDirectiveNoFileComp
}
