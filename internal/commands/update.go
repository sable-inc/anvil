package commands

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/sable-inc/anvil/internal/version"
)

const goModule = "github.com/sable-inc/anvil/cmd/anvil"

func newUpdateCmd() *cobra.Command {
	var ref string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update anvil to the latest version",
		Long: `Update anvil to the latest version using go install.

By default, installs the latest tagged release. Use --ref to install
a specific version, branch, or commit.`,
		Example: `  # Update to latest release
  anvil update

  # Update to latest main (bleeding edge)
  anvil update --ref main

  # Pin to a specific version
  anvil update --ref v0.2.0`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			w := cmd.OutOrStdout()

			// Check that go is available.
			goPath, err := exec.LookPath("go")
			if err != nil {
				return fmt.Errorf("go is not installed — download a binary from https://github.com/sable-inc/anvil/releases instead")
			}

			_, _ = fmt.Fprintf(w, "Current: %s\n", version.Info())

			target := "@latest"
			if ref != "" {
				target = "@" + ref
			}

			_, _ = fmt.Fprintf(w, "Installing %s%s ...\n", goModule, target)

			//nolint:gosec // ref is user-provided but only used as a go install version arg
			install := exec.Command(goPath, "install", goModule+target)
			install.Stdout = os.Stdout
			install.Stderr = os.Stderr
			// Inherit GOPRIVATE so private repos work.
			install.Env = appendGoPrivate(os.Environ())

			if err := install.Run(); err != nil {
				return fmt.Errorf("go install failed: %w", err)
			}

			// Run the new binary to get its version.
			newBin, err := exec.LookPath("anvil")
			if err == nil {
				out, verErr := exec.Command(newBin, "version").Output() //nolint:gosec // path from LookPath
				if verErr == nil {
					_, _ = fmt.Fprintf(w, "Updated: %s", strings.TrimSpace(string(out)))
					_, _ = fmt.Fprintln(w)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&ref, "ref", "", "Version, branch, or commit to install (default: latest tag)")
	return cmd
}

// appendGoPrivate ensures GOPRIVATE includes sable-inc for private repo access.
func appendGoPrivate(env []string) []string {
	for i, e := range env {
		if strings.HasPrefix(e, "GOPRIVATE=") {
			val := strings.TrimPrefix(e, "GOPRIVATE=")
			if !strings.Contains(val, "github.com/sable-inc") {
				env[i] = "GOPRIVATE=" + val + ",github.com/sable-inc/*"
			}
			return env
		}
	}
	return append(env, "GOPRIVATE=github.com/sable-inc/*")
}
