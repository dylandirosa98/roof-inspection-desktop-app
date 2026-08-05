package resources

import (
	"os"
	"path/filepath"
)

// Path prefers files installed beside the executable and falls back to the
// working directory for source-based development.
func Path(parts ...string) string {
	if executable, err := os.Executable(); err == nil {
		candidate := filepath.Join(append([]string{filepath.Dir(executable)}, parts...)...)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return filepath.Join(parts...)
}
