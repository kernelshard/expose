package version

var (
	Version   = "v0.6.1"
	GitCommit = "741ec5d"
	BuildDate = "2026-10-31"
)

// GetVersion returns just the version string
func GetVersion() string {
	if Version == "dev" {
		return "dev (unreleased)"
	}
	return Version
}

// GetFullVersion returns version with build metadata (for Cobra)
// Note: Cobra automatically prefixes with "expose version"
func GetFullVersion() string {
	return GetVersion() + " (commit: " + GitCommit + ", built: " + BuildDate + ")"
}
