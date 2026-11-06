package version

import (
	"strings"
	"testing"
)

func TestGetVersion(t *testing.T) {
	v := GetVersion()
	if !strings.Contains(v, Version) {
		t.Errorf("GetVersion mismatch expected %s in:%s", Version, v)
	}
}

func TestGetVersion_Dev(t *testing.T) {
	orig := Version
	t.Cleanup(func() { Version = orig })
	Version = "dev"
	if got := GetVersion(); got != "dev (unreleased)" {
		t.Errorf("expected 'dev (unreleased)', got: %s", got)
	}
}

func TestGetFullVersion(t *testing.T) {
	fullVersion := GetFullVersion()
	if !strings.Contains(fullVersion, GitCommit) {
		t.Errorf("GetFullVersion call expected %s in %s", GitCommit, fullVersion)
	}
	if !strings.Contains(fullVersion, BuildDate) {
		t.Errorf("GetFullVersion call expected %s in %s", BuildDate, fullVersion)
	}
}
