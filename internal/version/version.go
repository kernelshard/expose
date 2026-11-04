package version

var (
	Version   = "v0.6.2"
	GitCommit = "8974706"
	BuildDate = "2026-11-04"
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
