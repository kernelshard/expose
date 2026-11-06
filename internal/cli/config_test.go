package cli

import (
	"os"
	"strings"
	"testing"

	"github.com/kernelshard/expose/internal/config"
)

func TestConfigCmd_Structure(t *testing.T) {
	cmd := newConfigCmd()
	if cmd.Name() != "config" {
		t.Errorf("expected command name 'config', got %q", cmd.Name())
	}

	subCommands := cmd.Commands()
	if len(subCommands) != 2 {
		t.Fatalf("expected 2 subcommands, got %d", len(subCommands))
	}
}

func TestConfigListCmd(t *testing.T) {
	t.Run("fails when config does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		prevDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer os.Chdir(prevDir)

		cmd := newConfigListCmd()
		err := cmd.RunE(cmd, nil)
		if err == nil || !strings.Contains(err.Error(), "config not found") {
			t.Fatalf("expected 'config not found' error, got: %v", err)
		}
	})

	t.Run("succeeds when config exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		prevDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer os.Chdir(prevDir)

		_, err := config.Init()
		if err != nil {
			t.Fatalf("failed to init config: %v", err)
		}

		cmd := newConfigListCmd()
		if err := cmd.RunE(cmd, nil); err != nil {
			t.Fatalf("expected config list to succeed, got: %v", err)
		}
	})
}

func TestConfigGetCmd(t *testing.T) {
	t.Run("fails when config does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		prevDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer os.Chdir(prevDir)

		cmd := newConfigGetCmd()
		err := cmd.RunE(cmd, []string{"port"})
		if err == nil || !strings.Contains(err.Error(), "config not found") {
			t.Fatalf("expected 'config not found' error, got: %v", err)
		}
	})

	t.Run("succeeds for valid key and fails for invalid key", func(t *testing.T) {
		tmpDir := t.TempDir()
		prevDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer os.Chdir(prevDir)

		_, err := config.Init()
		if err != nil {
			t.Fatalf("failed to init config: %v", err)
		}

		cmd := newConfigGetCmd()

		// Valid key
		if err := cmd.RunE(cmd, []string{"port"}); err != nil {
			t.Fatalf("expected get 'port' to succeed, got: %v", err)
		}

		// Invalid key
		if err := cmd.RunE(cmd, []string{"nonexistent"}); err == nil {
			t.Fatal("expected get 'nonexistent' to fail, got nil")
		}
	})
}
