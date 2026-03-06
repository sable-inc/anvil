package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/sable-inc/anvil/internal/hyperdx"
)

// registerHyperDXTools adds HyperDX observability tools to the MCP server.
func registerHyperDXTools(s *server.MCPServer, h *Handler) {
	if h.hdx == nil {
		return
	}

	s.AddTool(mcp.NewTool("hdx_search_events",
		mcp.WithDescription(
			"Search HyperDX logs and spans with aggregation. "+
				"Use this to investigate errors, check latency percentiles, and monitor service health. "+
				"query uses HyperDX search syntax: 'level:err', 'service:sable-api', 'level:err AND service:api'. "+
				"Returns aggregated time-series data, not raw log lines.",
		),
		mcp.WithString("query", mcp.Description("HyperDX search filter (e.g. 'level:err', 'service:sable-api'). Leave empty for all events.")),
		mcp.WithString("agg", mcp.Description("Aggregation function"), mcp.Enum(
			"count", "avg", "sum", "min", "max",
			"p50", "p90", "p95", "p99", "count_distinct",
		)),
		mcp.WithString("field", mcp.Description("Field to aggregate on (e.g. 'duration'). Not required for count.")),
		mcp.WithString("group_by", mcp.Description("Comma-separated fields to group by (e.g. 'service,level')")),
		mcp.WithString("time_range", mcp.Description("Time range shorthand: 5m, 15m, 1h, 6h, 1d, 7d, 30d (default: 1h)")),
		mcp.WithString("granularity", mcp.Description("Time bucket size (e.g. '1 minute', '5 minute', '1 hour', '1 day'). Omit for a single aggregated value.")),
	), h.hdxSearchEvents)

	s.AddTool(mcp.NewTool("hdx_query_metrics",
		mcp.WithDescription(
			"Query HyperDX metrics (counters, gauges, histograms). "+
				"Use metric_data_type to specify the metric kind.",
		),
		mcp.WithString("query", mcp.Description("HyperDX search filter to narrow metrics.")),
		mcp.WithString("agg", mcp.Required(), mcp.Description("Aggregation function"), mcp.Enum(
			"avg", "sum", "min", "max", "count",
			"p50", "p90", "p95", "p99",
			"avg_rate", "sum_rate", "min_rate", "max_rate",
			"p50_rate", "p90_rate", "p95_rate", "p99_rate",
		)),
		mcp.WithString("field", mcp.Required(), mcp.Description("Metric name to query")),
		mcp.WithString("metric_data_type", mcp.Required(), mcp.Description("Metric type"), mcp.Enum("Sum", "Gauge", "Histogram")),
		mcp.WithString("group_by", mcp.Description("Comma-separated fields to group by")),
		mcp.WithString("time_range", mcp.Description("Time range shorthand: 5m, 15m, 1h, 6h, 1d, 7d, 30d (default: 1h)")),
		mcp.WithString("granularity", mcp.Description("Time bucket size (e.g. '1 minute', '1 hour'). Omit for a single aggregated value.")),
	), h.hdxQueryMetrics)

	s.AddTool(mcp.NewTool("hdx_recent_logs",
		mcp.WithDescription(
			"Get recent log messages grouped by content. "+
				"Returns the most frequent log messages with counts — use this to quickly understand what a service is doing. "+
				"Equivalent to grouping by the log body field. "+
				"Tip: combine with level filter (e.g. query='level:err') to see only errors.",
		),
		mcp.WithString("service", mcp.Description("Service name to filter (e.g. 'sable-api', 'sable-platform'). Leave empty for all services.")),
		mcp.WithString("query", mcp.Description("Additional HyperDX search filter (e.g. 'level:err', 'span_name:deploy*'). Combined with service filter.")),
		mcp.WithString("time_range", mcp.Description("Time range: 5m, 15m, 1h, 6h, 1d, 7d (default: 1h)")),
	), h.hdxRecentLogs)

	s.AddTool(mcp.NewTool("hdx_error_details",
		mcp.WithDescription(
			"Get a breakdown of errors across all services. "+
				"Returns error messages, which service they come from, and the operation (span) that failed. "+
				"Use this as the first tool when investigating production issues.",
		),
		mcp.WithString("service", mcp.Description("Filter to a specific service (e.g. 'sable-api'). Leave empty for all services.")),
		mcp.WithString("time_range", mcp.Description("Time range: 5m, 15m, 1h, 6h, 1d, 7d (default: 1h)")),
	), h.hdxErrorDetails)

	s.AddTool(mcp.NewTool("hdx_list_dashboards",
		mcp.WithDescription("Lists all HyperDX dashboards. Returns dashboard names, IDs, and tags."),
	), h.hdxListDashboards)

	s.AddTool(mcp.NewTool("hdx_get_dashboard",
		mcp.WithDescription(
			"Gets a specific HyperDX dashboard with all its chart definitions. "+
				"Use hdx_list_dashboards first to find valid IDs.",
		),
		mcp.WithString("dashboard_id", mcp.Required(), mcp.Description("Dashboard ID")),
	), h.hdxGetDashboard)

	s.AddTool(mcp.NewTool("hdx_list_alerts",
		mcp.WithDescription("Lists all configured HyperDX alerts with their thresholds and channels."),
	), h.hdxListAlerts)
}

func (h *Handler) hdxSearchEvents(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return h.hdxQuerySeries(ctx, req, "events", "")
}

func (h *Handler) hdxQueryMetrics(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	metricType := optString(req, "metric_data_type", "")
	if metricType == "" {
		return errResult("metric_data_type is required (Sum, Gauge, or Histogram)")
	}
	return h.hdxQuerySeries(ctx, req, "metrics", metricType)
}

func (h *Handler) hdxQuerySeries(ctx context.Context, req mcp.CallToolRequest, dataSource, metricType string) (*mcp.CallToolResult, error) {
	agg := optString(req, "agg", "count")
	field := optString(req, "field", "")
	query := optString(req, "query", "")
	groupByStr := optString(req, "group_by", "")
	timeRange := optString(req, "time_range", "1h")
	granularity := optString(req, "granularity", "")

	startMs, endMs := parseTimeRange(timeRange)

	groupBy := make([]string, 0)
	if groupByStr != "" {
		for _, g := range strings.Split(groupByStr, ",") {
			if trimmed := strings.TrimSpace(g); trimmed != "" {
				groupBy = append(groupBy, trimmed)
			}
		}
	}

	series := map[string]any{
		"dataSource": dataSource,
		"aggFn":      agg,
		"field":      field,
		"where":      query,
		"groupBy":    groupBy,
	}
	if metricType != "" {
		series["metricDataType"] = metricType
	}

	body := map[string]any{
		"series":    []any{series},
		"startTime": startMs,
		"endTime":   endMs,
	}
	if granularity != "" {
		body["granularity"] = granularity
	}

	resp, err := h.hdx.Post(ctx, "/api/v1/charts/series", body)
	if err != nil {
		return hdxErrResult("HyperDX query failed", err)
	}

	summary := fmt.Sprintf("Query: %s %s", agg, dataSource)
	if field != "" {
		summary += " on " + field
	}
	if query != "" {
		summary += " where " + query
	}
	if len(groupBy) > 0 {
		summary += " grouped by " + strings.Join(groupBy, ", ")
	}
	summary += " (" + timeRange + ")"
	if granularity != "" {
		summary += " granularity=" + granularity
	}

	// Parse data to provide a count for context.
	var parsed struct {
		Data []json.RawMessage `json:"data"`
	}
	rowCount := "unknown"
	if json.Unmarshal(resp, &parsed) == nil {
		rowCount = strconv.Itoa(len(parsed.Data))
	}

	text := fmt.Sprintf("%s\nRows: %s\n\n%s", summary, rowCount, string(resp))
	return mcp.NewToolResultText(text), nil
}

func (h *Handler) hdxRecentLogs(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := buildServiceQuery(optString(req, "service", ""), optString(req, "query", ""))
	timeRange := optString(req, "time_range", "1h")
	startMs, endMs := parseTimeRange(timeRange)

	body := map[string]any{
		"series": []any{
			map[string]any{
				"dataSource": "events",
				"aggFn":      "count",
				"field":      "",
				"where":      query,
				"groupBy":    []string{"body"},
			},
		},
		"startTime": startMs,
		"endTime":   endMs,
	}

	resp, err := h.hdx.Post(ctx, "/api/v1/charts/series", body)
	if err != nil {
		return hdxErrResult("Failed to fetch recent logs", err)
	}
	return formatLogResults(resp, "Recent logs", query, timeRange)
}

func (h *Handler) hdxErrorDetails(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	service := optString(req, "service", "")
	timeRange := optString(req, "time_range", "1h")
	startMs, endMs := parseTimeRange(timeRange)

	baseQuery := buildServiceQuery(service, "level:err")

	// Two series: group by body and group by span_name for a complete picture.
	bodyQuery := map[string]any{
		"series": []any{
			map[string]any{
				"dataSource": "events",
				"aggFn":      "count",
				"field":      "",
				"where":      baseQuery,
				"groupBy":    []string{"body", "service"},
			},
		},
		"startTime": startMs,
		"endTime":   endMs,
	}
	spanQuery := map[string]any{
		"series": []any{
			map[string]any{
				"dataSource": "events",
				"aggFn":      "count",
				"field":      "",
				"where":      baseQuery,
				"groupBy":    []string{"span_name", "service"},
			},
		},
		"startTime": startMs,
		"endTime":   endMs,
	}

	bodyResp, err := h.hdx.Post(ctx, "/api/v1/charts/series", bodyQuery)
	if err != nil {
		return hdxErrResult("Failed to fetch error details", err)
	}
	spanResp, err := h.hdx.Post(ctx, "/api/v1/charts/series", spanQuery)
	if err != nil {
		return hdxErrResult("Failed to fetch error spans", err)
	}

	header := fmt.Sprintf("Error details (%s)", timeRange)
	if service != "" {
		header += " service=" + service
	}

	text := header + "\n\n--- Errors by message ---\n" + string(bodyResp) +
		"\n\n--- Errors by operation ---\n" + string(spanResp)
	return mcp.NewToolResultText(text), nil
}

// buildServiceQuery combines a service filter and an optional extra query.
func buildServiceQuery(service, extra string) string {
	parts := make([]string, 0, 2)
	if service != "" {
		parts = append(parts, "service:"+service)
	}
	if extra != "" {
		parts = append(parts, extra)
	}
	return strings.Join(parts, " AND ")
}

// formatLogResults formats a chart/series response into a readable log summary.
func formatLogResults(resp json.RawMessage, header, query, timeRange string) (*mcp.CallToolResult, error) {
	var parsed struct {
		Data []json.RawMessage `json:"data"`
	}
	rowCount := "unknown"
	if json.Unmarshal(resp, &parsed) == nil {
		rowCount = strconv.Itoa(len(parsed.Data))
	}

	summary := header + " (" + timeRange + ")"
	if query != "" {
		summary += " where " + query
	}
	text := fmt.Sprintf("%s\nUnique messages: %s\n\n%s", summary, rowCount, string(resp))
	return mcp.NewToolResultText(text), nil
}

func (h *Handler) hdxListDashboards(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	resp, err := h.hdx.Get(ctx, "/api/v1/dashboards")
	if err != nil {
		return hdxErrResult("Failed to list HyperDX dashboards", err)
	}
	return mcp.NewToolResultText(string(resp)), nil
}

func (h *Handler) hdxGetDashboard(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("dashboard_id")
	if err != nil {
		return errResult("dashboard_id is required. Use hdx_list_dashboards to find valid IDs.")
	}
	resp, err := h.hdx.Get(ctx, "/api/v1/dashboards/"+id)
	if err != nil {
		return hdxErrResult("Failed to get HyperDX dashboard", err)
	}
	return mcp.NewToolResultText(string(resp)), nil
}

func (h *Handler) hdxListAlerts(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	resp, err := h.hdx.Get(ctx, "/api/v1/alerts")
	if err != nil {
		return hdxErrResult("Failed to list HyperDX alerts", err)
	}
	return mcp.NewToolResultText(string(resp)), nil
}

// hdxErrResult returns an actionable error for HyperDX API failures.
func hdxErrResult(context string, err error) (*mcp.CallToolResult, error) {
	if hyperdx.IsUnauthorized(err) {
		return errResult("%s: unauthorized. Run 'anvil settings set-hyperdx <key>' with a valid HyperDX API key.", context)
	}
	return errResult("%s: %v", context, err)
}

// parseTimeRange converts a shorthand like "1h", "6h", "1d" into start/end epoch milliseconds.
// Supported: 5m, 15m, 30m, 1h, 6h, 12h, 1d, 7d, 30d. Defaults to 1h for unrecognized input.
func parseTimeRange(s string) (startMs, endMs int64) {
	now := time.Now()
	endMs = now.UnixMilli()

	durations := map[string]time.Duration{
		"5m":  5 * time.Minute,
		"15m": 15 * time.Minute,
		"30m": 30 * time.Minute,
		"1h":  1 * time.Hour,
		"6h":  6 * time.Hour,
		"12h": 12 * time.Hour,
		"1d":  24 * time.Hour,
		"7d":  7 * 24 * time.Hour,
		"30d": 30 * 24 * time.Hour,
	}

	d, ok := durations[s]
	if !ok {
		d = time.Hour
	}
	startMs = now.Add(-d).UnixMilli()
	return startMs, endMs
}
