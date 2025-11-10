package version

var (
	Version   = "v0.1.2"
	GitCommit = "d30c483"
	BuildDate = "2025-11-10"
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
