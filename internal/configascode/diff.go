package configascode

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

// FieldDiff represents a change in a single config field.
type FieldDiff struct {
	Path   string `json:"path"`
	Local  any    `json:"local"`
	Remote any    `json:"remote"`
}

// DiffResult holds the structured diff between local and remote configs.
type DiffResult struct {
	Changed []FieldDiff `json:"changed,omitempty"`
	Added   []FieldDiff `json:"added,omitempty"`
	Removed []FieldDiff `json:"removed,omitempty"`
}

// HasChanges returns true if there are any differences.
func (d *DiffResult) HasChanges() bool {
	return len(d.Changed) > 0 || len(d.Added) > 0 || len(d.Removed) > 0
}

// Diff compares two AgentConfigs and returns the differences.
// It normalizes both to JSON maps for a generic field-by-field comparison.
func Diff(local, remote *AgentConfig) (*DiffResult, error) {
	localMap, err := toMap(local)
	if err != nil {
		return nil, fmt.Errorf("converting local config: %w", err)
	}
	remoteMap, err := toMap(remote)
	if err != nil {
		return nil, fmt.Errorf("converting remote config: %w", err)
	}

	result := &DiffResult{}
	diffMaps("", localMap, remoteMap, result)
	return result, nil
}

// WriteDiff renders a human-readable diff to the writer.
func WriteDiff(w io.Writer, result *DiffResult) error {
	if !result.HasChanges() {
		_, err := fmt.Fprintln(w, "No differences found.")
		return err
	}

	if len(result.Changed) > 0 {
		if _, err := fmt.Fprintln(w, "Changed:"); err != nil {
			return err
		}
		for _, d := range result.Changed {
			if _, err := fmt.Fprintf(w, "  %s: %v → %v\n", d.Path, formatValue(d.Local), formatValue(d.Remote)); err != nil {
				return err
			}
		}
	}

	if len(result.Added) > 0 {
		if _, err := fmt.Fprintln(w, "Added (local only):"); err != nil {
			return err
		}
		for _, d := range result.Added {
			if _, err := fmt.Fprintf(w, "  + %s: %v\n", d.Path, formatValue(d.Local)); err != nil {
				return err
			}
		}
	}

	if len(result.Removed) > 0 {
		if _, err := fmt.Fprintln(w, "Removed (remote only):"); err != nil {
			return err
		}
		for _, d := range result.Removed {
			if _, err := fmt.Fprintf(w, "  - %s: %v\n", d.Path, formatValue(d.Remote)); err != nil {
				return err
			}
		}
	}

	return nil
}

func toMap(v any) (map[string]any, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func diffMaps(prefix string, local, remote map[string]any, result *DiffResult) {
	allKeys := make(map[string]struct{})
	for k := range local {
		allKeys[k] = struct{}{}
	}
	for k := range remote {
		allKeys[k] = struct{}{}
	}

	sorted := make([]string, 0, len(allKeys))
	for k := range allKeys {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)

	for _, key := range sorted {
		path := key
		if prefix != "" {
			path = prefix + "." + key
		}

		lv, lok := local[key]
		rv, rok := remote[key]

		switch {
		case lok && !rok:
			result.Added = append(result.Added, FieldDiff{Path: path, Local: lv})
		case !lok && rok:
			result.Removed = append(result.Removed, FieldDiff{Path: path, Remote: rv})
		default:
			// Both present — recurse into maps or compare values.
			lm, lIsMap := lv.(map[string]any)
			rm, rIsMap := rv.(map[string]any)
			if lIsMap && rIsMap {
				diffMaps(path, lm, rm, result)
			} else if !jsonEqual(lv, rv) {
				result.Changed = append(result.Changed, FieldDiff{Path: path, Local: lv, Remote: rv})
			}
		}
	}
}

func jsonEqual(a, b any) bool {
	aj, _ := json.Marshal(a)
	bj, _ := json.Marshal(b)
	return string(aj) == string(bj)
}

func formatValue(v any) string {
	if v == nil {
		return "null"
	}
	switch val := v.(type) {
	case string:
		if len(val) > 60 {
			return fmt.Sprintf("%q...", val[:57])
		}
		return fmt.Sprintf("%q", val)
	case bool, float64, int:
		return fmt.Sprintf("%v", val)
	default:
		data, err := json.Marshal(val)
		if err != nil {
			return fmt.Sprintf("%v", val)
		}
		s := string(data)
		if len(s) > 60 {
			return s[:57] + "..."
		}
		return s
	}
}

// SummaryLine returns a one-line summary of the diff.
func SummaryLine(result *DiffResult) string {
	if !result.HasChanges() {
		return "configs are identical"
	}
	parts := make([]string, 0, 3)
	if n := len(result.Changed); n > 0 {
		parts = append(parts, fmt.Sprintf("%d changed", n))
	}
	if n := len(result.Added); n > 0 {
		parts = append(parts, fmt.Sprintf("%d added", n))
	}
	if n := len(result.Removed); n > 0 {
		parts = append(parts, fmt.Sprintf("%d removed", n))
	}
	return strings.Join(parts, ", ")
}
