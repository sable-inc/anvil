package output

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// Table builds columnar text output.
type Table struct {
	headers []string
	rows    [][]string
}

// NewTable creates a table with the given column headers.
func NewTable(headers ...string) *Table {
	return &Table{headers: headers}
}

// AddRow appends a row of values. Values should match the header count.
func (t *Table) AddRow(values ...string) {
	t.rows = append(t.rows, values)
}

// Render writes the table to w using aligned tab-separated columns.
func (t *Table) Render(w io.Writer) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	// Header row.
	upper := make([]string, len(t.headers))
	for i, h := range t.headers {
		upper[i] = strings.ToUpper(h)
	}
	if _, err := fmt.Fprintln(tw, strings.Join(upper, "\t")); err != nil {
		return err
	}

	// Data rows.
	for _, row := range t.rows {
		if _, err := fmt.Fprintln(tw, strings.Join(row, "\t")); err != nil {
			return err
		}
	}

	return tw.Flush()
}

// Write dispatches output based on format.
// JSON and YAML serialize data; table format uses the provided Table.
func Write(w io.Writer, format string, data any, table *Table) error {
	switch format {
	case "json":
		return (&jsonFormatter{}).Format(w, data)
	case "yaml":
		return (&yamlFormatter{}).Format(w, data)
	default:
		return table.Render(w)
	}
}
