package plugin

import (
	"os"
	"strings"
)

// For all mounts, map the `~` symbol to `$HOME` and expand it.
func expandConfigPath(mounts []string) []string {
	expandedPaths := make([]string, len(mounts))
	for i, mount := range mounts {
		// As `~` is not handled by `ExpandEnv()`, replace it with `$HOME`.
		mount = strings.Replace(mount, "~", "$HOME", 1)
		expandedPaths[i] = os.ExpandEnv(mount)
	}
	return expandedPaths
}
