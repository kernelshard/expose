package version

var (
	Version   = "v0.6.0"
	GitCommit = "5ab3678"
	BuildDate = "2026-10-30"
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
