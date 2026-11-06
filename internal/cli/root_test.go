package cli

import (
	"testing"
)

func TestExecute(t *testing.T) {
	rootCmd.SetArgs([]string{"--help"})
	if err := Execute(); err != nil {
		t.Fatalf("expected Execute() with --help to succeed, got: %v", err)
	}
}

