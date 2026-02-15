// Package version provides build-time version information injected via ldflags.
package version

import "fmt"

// These variables are set at build time via -ldflags.
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

// Info returns a formatted version string.
func Info() string {
	return fmt.Sprintf("anvil %s (commit: %s, built: %s)", Version, Commit, Date)
}
