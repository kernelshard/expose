package cli

import (
	"os"
	"testing"
)

func TestInitCmd(t *testing.T) {
	cmd := newInitCmd()
	if cmd == nil {
		t.Fatal("expected newInitCmd to return non-nil command")
	}
	if cmd.Name() != "init" {
		t.Errorf("expected command name 'init', got %q", cmd.Name())
	}

	t.Run("successfully creates config and rejects duplicate", func(t *testing.T) {
		tmpDir := t.TempDir()
		prevDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer os.Chdir(prevDir)

		// First execution creates .expose.yml
		err := cmd.RunE(cmd, nil)
		if err != nil {
			t.Fatalf("expected init to succeed, got: %v", err)
		}

		// Second execution should fail because config already exists
		err = cmd.RunE(cmd, nil)
		if err == nil {
			t.Fatal("expected second init to fail with 'config already exists', got nil")
		}
	})
}

