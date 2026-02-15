package commands

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/sable-inc/anvil/internal/output"
)

// ForgeVersion represents a tagged release of the forge (agentkit) repo.
type ForgeVersion struct {
	Name string `json:"name" yaml:"name"`
	SHA  string `json:"sha" yaml:"sha"`
}

// ForgeBranch represents a branch in the forge repo.
type ForgeBranch struct {
	Name string `json:"name" yaml:"name"`
	SHA  string `json:"sha" yaml:"sha"`
}

// ForgeCommit represents a commit in the forge repo.
type ForgeCommit struct {
	SHA     string `json:"sha" yaml:"sha"`
	Message string `json:"message" yaml:"message"`
	Author  string `json:"author" yaml:"author"`
	Date    string `json:"date" yaml:"date"`
}

func newForgeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "forge",
		Short: "Manage forge (agentkit) versions",
		Long:  "List versions, branches, commits, and validate refs for the forge repository.",
	}

	cmd.AddCommand(newForgeVersionsCmd())
	cmd.AddCommand(newForgeBranchesCmd())
	cmd.AddCommand(newForgeCommitsCmd())
	cmd.AddCommand(newForgeValidateCmd())
	return cmd
}

func newForgeVersionsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "versions",
		Short: "List forge versions (tags)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}

			var resp struct {
				Versions []ForgeVersion `json:"versions"`
			}
			if err := client.Get(cmd.Context(), "/forge-versions", &resp); err != nil {
				return err
			}

			t := output.NewTable("Name", "SHA")
			for _, v := range resp.Versions {
				t.AddRow(v.Name, truncateSHA(v.SHA))
			}
			return output.Write(a.Out, a.Format, resp.Versions, t)
		},
	}
}

func newForgeBranchesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "branches",
		Short: "List forge branches",
		RunE: func(cmd *cobra.Command, _ []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}

			var resp struct {
				Branches []ForgeBranch `json:"branches"`
			}
			if err := client.Get(cmd.Context(), "/forge-branches", &resp); err != nil {
				return err
			}

			t := output.NewTable("Name", "SHA")
			for _, b := range resp.Branches {
				t.AddRow(b.Name, truncateSHA(b.SHA))
			}
			return output.Write(a.Out, a.Format, resp.Branches, t)
		},
	}
}

func newForgeCommitsCmd() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "commits <branch>",
		Short: "List recent commits on a branch",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}

			path := "/forge-commits?branch=" + args[0]
			if limit > 0 {
				path += "&limit=" + strconv.Itoa(limit)
			}

			var resp struct {
				Commits []ForgeCommit `json:"commits"`
			}
			if err := client.Get(cmd.Context(), path, &resp); err != nil {
				return err
			}

			t := output.NewTable("SHA", "Author", "Date", "Message")
			for _, c := range resp.Commits {
				t.AddRow(truncateSHA(c.SHA), c.Author, c.Date, truncateMsg(c.Message))
			}
			return output.Write(a.Out, a.Format, resp.Commits, t)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 0, "Max number of commits to show")
	return cmd
}

func newForgeValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate <ref>",
		Short: "Validate a git ref (branch, tag, or SHA)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}

			body := map[string]any{"ref": args[0]}
			var resp struct {
				Valid       bool    `json:"valid" yaml:"valid"`
				ResolvedSHA *string `json:"resolvedSha" yaml:"resolvedSha"`
				Error       *string `json:"error" yaml:"error"`
			}
			if err := client.Post(cmd.Context(), "/forge-validate-ref", body, &resp); err != nil {
				return err
			}

			if resp.Valid {
				_, err = fmt.Fprintf(a.Out, "Valid ref %q → %s\n", args[0], ptrStr(resp.ResolvedSHA))
			} else {
				_, err = fmt.Fprintf(a.Out, "Invalid ref %q: %s\n", args[0], ptrStr(resp.Error))
			}
			return err
		},
	}
}

func truncateSHA(sha string) string {
	if len(sha) > 8 {
		return sha[:8]
	}
	return sha
}

func truncateMsg(msg string) string {
	if len(msg) > 60 {
		return msg[:57] + "..."
	}
	return msg
}
