package version_test

import (
	"strings"
	"testing"

	"github.com/sable-inc/anvil/internal/version"
)

func TestInfo(t *testing.T) {
	info := version.Info()

	if !strings.Contains(info, "anvil") {
		t.Errorf("Info() = %q, want to contain %q", info, "anvil")
	}
	if !strings.Contains(info, version.Version) {
		t.Errorf("Info() = %q, want to contain version %q", info, version.Version)
	}
}
