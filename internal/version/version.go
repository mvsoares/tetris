package version

import "fmt"

var (
	// Version is the current release version of Tetris.
	Version = "1.0.0"

	// Commit is the git commit hash set at build time.
	Commit = "none"

	// Date is the build date timestamp set at build time.
	Date = "unknown"
)

// Full returns a human-readable formatted version string.
func Full() string {
	if Commit != "none" && Commit != "" {
		return fmt.Sprintf("v%s (%s, %s)", Version, Commit, Date)
	}
	return fmt.Sprintf("v%s", Version)
}
