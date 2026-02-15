// Package output provides output formatting for the Anvil CLI.
package output

import (
	"encoding/json"
	"io"

	"gopkg.in/yaml.v3"
)

// Formatter renders a value to a writer in a specific format.
type Formatter interface {
	Format(w io.Writer, v any) error
}

// New returns a Formatter for the given format string.
// Supported formats: "json", "yaml", "table".
// Table falls back to JSON until Phase 3 adds the table renderer.
func New(format string) Formatter {
	switch format {
	case "json":
		return &jsonFormatter{}
	case "yaml":
		return &yamlFormatter{}
	case "table":
		return &tableFormatter{}
	default:
		return &jsonFormatter{}
	}
}

type jsonFormatter struct{}

func (f *jsonFormatter) Format(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

type yamlFormatter struct{}

func (f *yamlFormatter) Format(w io.Writer, v any) error {
	data, err := yaml.Marshal(v)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

// tableFormatter falls back to JSON for unstructured data.
// Commands should prefer output.Write() with a Table for proper table rendering.
type tableFormatter struct{}

func (f *tableFormatter) Format(w io.Writer, v any) error {
	return (&jsonFormatter{}).Format(w, v)
}
