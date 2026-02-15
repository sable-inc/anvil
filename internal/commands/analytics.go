package commands

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/sable-inc/anvil/internal/output"
)

func newAnalyticsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analytics",
		Short: "View session analytics",
	}

	cmd.AddCommand(newAnalyticsSessionsCmd())
	cmd.AddCommand(newAnalyticsStagesCmd())
	return cmd
}

func newAnalyticsSessionsCmd() *cobra.Command {
	var timeRange, groupBy string

	cmd := &cobra.Command{
		Use:   "sessions",
		Short: "View session statistics and time series",
		RunE: func(cmd *cobra.Command, _ []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}

			path := "/analytics/sessions"
			sep := "?"
			if a.OrgID != "" {
				path += sep + "orgId=" + a.OrgID
				sep = "&"
			}
			if timeRange != "" {
				path += sep + "timeRange=" + timeRange
				sep = "&"
			}
			if groupBy != "" {
				path += sep + "groupBy=" + groupBy
			}

			var resp struct {
				Stats struct {
					TotalSessions         int    `json:"totalSessions"`
					AvgSessionTimeMinutes float64 `json:"avgSessionTimeMinutes"`
					Growth                struct {
						Sessions string `json:"sessions"`
						AvgTime  string `json:"avgTime"`
					} `json:"growth"`
				} `json:"stats"`
				TimeSeries struct {
					Sessions []struct {
						Period string `json:"period"`
						Value  int    `json:"value"`
					} `json:"sessions"`
				} `json:"timeSeries"`
			}
			if err := client.Get(cmd.Context(), path, &resp); err != nil {
				return err
			}

			// For JSON/YAML, output raw response.
			if a.Format == "json" || a.Format == "yaml" {
				f := output.New(a.Format)
				return f.Format(a.Out, resp)
			}

			// Table: show summary stats and time series.
			if _, err := fmt.Fprintf(a.Out,
				"Total Sessions: %d  |  Avg Time: %.1f min  |  Growth: %s\n\n",
				resp.Stats.TotalSessions,
				resp.Stats.AvgSessionTimeMinutes,
				resp.Stats.Growth.Sessions,
			); err != nil {
				return err
			}

			t := output.NewTable("Period", "Sessions")
			for _, dp := range resp.TimeSeries.Sessions {
				t.AddRow(dp.Period, strconv.Itoa(dp.Value))
			}
			return t.Render(a.Out)
		},
	}

	cmd.Flags().StringVar(&timeRange, "range", "", "Time range (last7d|last30d|last90d|thisMonth|thisQuarter|thisYear)")
	cmd.Flags().StringVar(&groupBy, "group-by", "", "Group by (daily|weekly|monthly|quarterly)")
	return cmd
}

func newAnalyticsStagesCmd() *cobra.Command {
	var timeRange string

	cmd := &cobra.Command{
		Use:   "stages",
		Short: "View stage funnel analytics",
		RunE: func(cmd *cobra.Command, _ []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}

			path := "/analytics/stages"
			sep := "?"
			if a.OrgID != "" {
				path += sep + "orgId=" + a.OrgID
				sep = "&"
			}
			if timeRange != "" {
				path += sep + "timeRange=" + timeRange
			}

			var resp struct {
				Stages []struct {
					Stage      string  `json:"stage"`
					Count      int     `json:"count"`
					Percentage float64 `json:"percentage"`
				} `json:"stages"`
				TotalSessions int `json:"totalSessions"`
			}
			if err := client.Get(cmd.Context(), path, &resp); err != nil {
				return err
			}

			if a.Format == "json" || a.Format == "yaml" {
				f := output.New(a.Format)
				return f.Format(a.Out, resp)
			}

			if _, err := fmt.Fprintf(a.Out, "Total Sessions: %d\n\n", resp.TotalSessions); err != nil {
				return err
			}

			t := output.NewTable("Stage", "Count", "Percentage")
			for _, s := range resp.Stages {
				t.AddRow(s.Stage, strconv.Itoa(s.Count), fmt.Sprintf("%.1f%%", s.Percentage))
			}
			return t.Render(a.Out)
		},
	}

	cmd.Flags().StringVar(&timeRange, "range", "", "Time range (last7d|last30d|last90d|thisMonth|thisQuarter|thisYear)")
	return cmd
}
