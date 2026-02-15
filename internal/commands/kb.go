package commands

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/sable-inc/anvil/internal/output"
)

// KBItem mirrors the sable-api KnowledgeBaseItem response shape.
type KBItem struct {
	ID            int     `json:"id" yaml:"id"`
	OrgID         *int    `json:"orgId" yaml:"orgId"`
	Name          string  `json:"name" yaml:"name"`
	Type          string  `json:"type" yaml:"type"`
	Status        *string `json:"status" yaml:"status"`
	Enabled       *bool   `json:"enabled" yaml:"enabled"`
	SourceURL     *string `json:"sourceUrl" yaml:"sourceUrl"`
	PageTitle     *string `json:"pageTitle" yaml:"pageTitle"`
	ChunkCount    *int    `json:"chunkCount" yaml:"chunkCount"`
	WordCount     *int    `json:"wordCount" yaml:"wordCount"`
	FilePath      *string `json:"filePath" yaml:"filePath"`
	ParentItemID  *int    `json:"parentItemId" yaml:"parentItemId"`
	LastSyncedAt  *string `json:"lastSyncedAt" yaml:"lastSyncedAt"`
	LastCrawledAt *string `json:"lastCrawledAt" yaml:"lastCrawledAt"`
	CrawlError    *string `json:"crawlError" yaml:"crawlError"`
	CreatedAt     string  `json:"createdAt" yaml:"createdAt"`
	UpdatedAt     string  `json:"updatedAt" yaml:"updatedAt"`
}

func newKBCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "kb",
		Aliases: []string{"knowledge-base"},
		Short:   "Manage knowledge base",
	}

	cmd.AddCommand(newKBListCmd())
	cmd.AddCommand(newKBGetCmd())
	cmd.AddCommand(newKBSearchCmd())
	cmd.AddCommand(newKBImportURLCmd())
	cmd.AddCommand(newKBImportSitemapCmd())
	cmd.AddCommand(newKBSyncCmd())
	cmd.AddCommand(newKBSyncAllCmd())
	cmd.AddCommand(newKBCrawlCmd())
	cmd.AddCommand(newKBDeleteCmd())
	cmd.AddCommand(newKBJobCmd())
	return cmd
}

func newKBListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List knowledge base items",
		RunE: func(cmd *cobra.Command, _ []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}

			path := "/knowledge-base"
			if a.OrgID != "" {
				path += "?orgId=" + a.OrgID
			}

			var resp struct {
				Items []KBItem `json:"items"`
			}
			if err := client.Get(cmd.Context(), path, &resp); err != nil {
				return err
			}

			t := output.NewTable("ID", "Name", "Type", "Status", "Chunks", "Source URL")
			for _, item := range resp.Items {
				status := derefStr(item.Status, "-")
				chunks := "-"
				if item.ChunkCount != nil {
					chunks = strconv.Itoa(*item.ChunkCount)
				}
				src := "-"
				if item.SourceURL != nil {
					src = *item.SourceURL
					if len(src) > 50 {
						src = src[:47] + "..."
					}
				}
				t.AddRow(strconv.Itoa(item.ID), item.Name, item.Type, status, chunks, src)
			}
			return output.Write(a.Out, a.Format, resp.Items, t)
		},
	}
}

func newKBGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get knowledge base item details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}

			var resp struct {
				Item KBItem `json:"item"`
			}
			if err := client.Get(cmd.Context(), "/knowledge-base/"+args[0], &resp); err != nil {
				return err
			}

			item := resp.Item
			t := output.NewTable("Field", "Value")
			t.AddRow("ID", strconv.Itoa(item.ID))
			t.AddRow("Name", item.Name)
			t.AddRow("Type", item.Type)
			t.AddRow("Status", derefStr(item.Status, "-"))
			t.AddRow("Enabled", derefBoolStr(item.Enabled))
			t.AddRow("Chunks", derefIntStr(item.ChunkCount))
			t.AddRow("Words", derefIntStr(item.WordCount))
			if item.SourceURL != nil {
				t.AddRow("Source URL", *item.SourceURL)
			}
			if item.PageTitle != nil {
				t.AddRow("Page Title", *item.PageTitle)
			}
			t.AddRow("Last Synced", derefStr(item.LastSyncedAt, "-"))
			t.AddRow("Created", item.CreatedAt)
			t.AddRow("Updated", item.UpdatedAt)
			return output.Write(a.Out, a.Format, item, t)
		},
	}
}

func newKBSearchCmd() *cobra.Command {
	var topK int

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search the knowledge base",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}

			body := map[string]any{"query": args[0], "topK": topK}
			var resp struct {
				Query   string `json:"query"`
				Results []struct {
					ID       string         `json:"id"`
					Score    float64        `json:"score"`
					Content  string         `json:"content"`
					Metadata map[string]any `json:"metadata"`
				} `json:"results"`
			}
			if err := client.Post(cmd.Context(), "/knowledge-base/search", body, &resp); err != nil {
				return err
			}

			t := output.NewTable("ID", "Score", "Content")
			for _, r := range resp.Results {
				content := r.Content
				if len(content) > 80 {
					content = content[:77] + "..."
				}
				t.AddRow(r.ID, fmt.Sprintf("%.3f", r.Score), content)
			}
			return output.Write(a.Out, a.Format, resp, t)
		},
	}

	cmd.Flags().IntVar(&topK, "top-k", 5, "Number of results to return (1-20)")
	return cmd
}

func newKBImportURLCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "import-url <url>",
		Short: "Import a URL into the knowledge base",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}

			body := map[string]any{"url": args[0]}
			if name != "" {
				body["name"] = name
			}

			var resp struct {
				Item KBItem `json:"item"`
			}
			if err := client.Post(cmd.Context(), "/knowledge-base/url", body, &resp); err != nil {
				return err
			}

			item := resp.Item
			t := output.NewTable("Field", "Value")
			t.AddRow("ID", strconv.Itoa(item.ID))
			t.AddRow("Name", item.Name)
			t.AddRow("Type", item.Type)
			t.AddRow("Status", derefStr(item.Status, "-"))
			return output.Write(a.Out, a.Format, item, t)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Custom name for the item")
	return cmd
}

func newKBImportSitemapCmd() *cobra.Command {
	var (
		name     string
		maxURLs  int
		includes []string
		excludes []string
	)

	cmd := &cobra.Command{
		Use:   "import-sitemap <sitemap-url>",
		Short: "Import URLs from a sitemap",
		Long:  "Import knowledge base items from a sitemap. URLs are discovered, filtered, and processed asynchronously.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}

			body := map[string]any{
				"sitemapUrl": args[0],
				"name":       name,
			}

			filterOpts := map[string]any{}
			if len(includes) > 0 {
				filterOpts["includePatterns"] = includes
			}
			if len(excludes) > 0 {
				filterOpts["excludePatterns"] = excludes
			}
			if maxURLs > 0 {
				filterOpts["maxUrls"] = maxURLs
			}
			if len(filterOpts) > 0 {
				body["filterOptions"] = filterOpts
			}

			var resp struct {
				Item  KBItem `json:"item"`
				JobID string `json:"jobId"`
			}
			if err := client.Post(cmd.Context(), "/knowledge-base/sitemap", body, &resp); err != nil {
				return err
			}

			item := resp.Item
			t := output.NewTable("Field", "Value")
			t.AddRow("ID", strconv.Itoa(item.ID))
			t.AddRow("Name", item.Name)
			t.AddRow("Job ID", resp.JobID)
			t.AddRow("Status", derefStr(item.Status, "pending"))
			return output.Write(a.Out, a.Format, resp, t)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Name for the sitemap import (required)")
	cmd.Flags().IntVar(&maxURLs, "max-urls", 0, "Maximum URLs to process (default: 500)")
	cmd.Flags().StringSliceVar(&includes, "include", nil, "URL include patterns (e.g. '*/docs/*')")
	cmd.Flags().StringSliceVar(&excludes, "exclude", nil, "URL exclude patterns (e.g. '*/blog/*')")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newKBSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync <id>",
		Short: "Sync a knowledge base item to the vector store",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}

			var resp struct {
				Item       KBItem `json:"item"`
				Synced     bool   `json:"synced"`
				ChunkCount int    `json:"chunkCount"`
				WordCount  int    `json:"wordCount"`
			}
			if err := client.Post(cmd.Context(), "/knowledge-base/"+args[0]+"/sync", nil, &resp); err != nil {
				return err
			}

			t := output.NewTable("Field", "Value")
			t.AddRow("ID", strconv.Itoa(resp.Item.ID))
			t.AddRow("Name", resp.Item.Name)
			t.AddRow("Synced", fmt.Sprintf("%v", resp.Synced))
			t.AddRow("Chunks", strconv.Itoa(resp.ChunkCount))
			t.AddRow("Words", strconv.Itoa(resp.WordCount))
			return output.Write(a.Out, a.Format, resp, t)
		},
	}
}

func newKBSyncAllCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync-all",
		Short: "Sync all knowledge base items to the vector store",
		RunE: func(cmd *cobra.Command, _ []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}

			path := "/knowledge-base/sync-all"
			if a.OrgID != "" {
				path += "?orgId=" + a.OrgID
			}

			var resp struct {
				Total     int `json:"total"`
				Succeeded int `json:"succeeded"`
				Failed    int `json:"failed"`
				Deleted   int `json:"deleted"`
				Results   []struct {
					ID      int    `json:"id"`
					Success bool   `json:"success"`
					Error   string `json:"error,omitempty"`
				} `json:"results"`
			}
			if err := client.Post(cmd.Context(), path, nil, &resp); err != nil {
				return err
			}

			t := output.NewTable("Metric", "Value")
			t.AddRow("Total", strconv.Itoa(resp.Total))
			t.AddRow("Succeeded", strconv.Itoa(resp.Succeeded))
			t.AddRow("Failed", strconv.Itoa(resp.Failed))
			t.AddRow("Deleted", strconv.Itoa(resp.Deleted))
			return output.Write(a.Out, a.Format, resp, t)
		},
	}
}

func newKBCrawlCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "crawl <id>",
		Short: "Re-crawl a URL knowledge base item",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}

			var resp struct {
				Item       KBItem `json:"item"`
				Synced     bool   `json:"synced"`
				ChunkCount int    `json:"chunkCount"`
				WordCount  int    `json:"wordCount"`
			}
			if err := client.Post(cmd.Context(), "/knowledge-base/"+args[0]+"/crawl", nil, &resp); err != nil {
				return err
			}

			t := output.NewTable("Field", "Value")
			t.AddRow("ID", strconv.Itoa(resp.Item.ID))
			t.AddRow("Name", resp.Item.Name)
			t.AddRow("Synced", fmt.Sprintf("%v", resp.Synced))
			t.AddRow("Chunks", strconv.Itoa(resp.ChunkCount))
			t.AddRow("Words", strconv.Itoa(resp.WordCount))
			return output.Write(a.Out, a.Format, resp, t)
		},
	}
}

func newKBDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a knowledge base item",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}

			var resp struct {
				Success bool `json:"success"`
			}
			if err := client.Delete(cmd.Context(), "/knowledge-base/"+args[0], &resp); err != nil {
				return err
			}

			_, err = fmt.Fprintf(a.Out, "Knowledge base item %s deleted.\n", args[0])
			return err
		},
	}
}

func newKBJobCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "job <jobId>",
		Short: "Get sitemap import job status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}

			var resp struct {
				JobID    string `json:"jobId"`
				Status   string `json:"status"`
				Progress *struct {
					Stage          string `json:"stage"`
					Progress       int    `json:"progress"`
					Message        string `json:"message"`
					URLsDiscovered *int   `json:"urlsDiscovered"`
					URLsProcessed  *int   `json:"urlsProcessed"`
					URLsFailed     *int   `json:"urlsFailed"`
				} `json:"progress,omitempty"`
				Error *struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				} `json:"error,omitempty"`
			}
			if err := client.Get(cmd.Context(), "/knowledge-base/jobs/"+args[0], &resp); err != nil {
				return err
			}

			t := output.NewTable("Field", "Value")
			t.AddRow("Job ID", resp.JobID)
			t.AddRow("Status", resp.Status)
			if resp.Progress != nil {
				t.AddRow("Stage", resp.Progress.Stage)
				t.AddRow("Progress", fmt.Sprintf("%d%%", resp.Progress.Progress))
				t.AddRow("Message", resp.Progress.Message)
				if resp.Progress.URLsDiscovered != nil {
					t.AddRow("URLs Discovered", strconv.Itoa(*resp.Progress.URLsDiscovered))
				}
				if resp.Progress.URLsProcessed != nil {
					t.AddRow("URLs Processed", strconv.Itoa(*resp.Progress.URLsProcessed))
				}
				if resp.Progress.URLsFailed != nil {
					t.AddRow("URLs Failed", strconv.Itoa(*resp.Progress.URLsFailed))
				}
			}
			if resp.Error != nil {
				t.AddRow("Error Code", resp.Error.Code)
				t.AddRow("Error Message", resp.Error.Message)
			}
			return output.Write(a.Out, a.Format, resp, t)
		},
	}
}

// Helper functions for rendering optional fields.

func derefStr(p *string, fallback string) string {
	if p != nil {
		return *p
	}
	return fallback
}

func derefIntStr(p *int) string {
	if p != nil {
		return strconv.Itoa(*p)
	}
	return "-"
}

func derefBoolStr(p *bool) string {
	if p == nil {
		return "-"
	}
	if *p {
		return "true"
	}
	return "false"
}
