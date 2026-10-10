package cli

import (
	"os"
	"testing"
)

func TestTunnelCmd(t *testing.T) {
	cmd := newTunnelCmd()

	if cmd == nil {
		t.Fatal("newTunnelCmd returned nil")
	}

	if cmd.Name() != "tunnel" {
		t.Errorf("expected command name 'tunnel', got '%s'", cmd.Name())
	}

	// Check flag parsing
	flag := cmd.Flags().Lookup("port")
	if flag == nil {
		t.Error("port flag not defined")
	}

	if flag.Shorthand != "p" {
		t.Errorf("expected shorthand 'p' got %s", flag.Shorthand)
	}
}

func TestResolvePort(t *testing.T) {
	t.Run("defaults to 3000 when no config, args, or flags", func(t *testing.T) {
		tmpDir := t.TempDir()
		prevDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer os.Chdir(prevDir)

		cmd := newTunnelCmd()
		port, err := resolvePort(cmd, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if port != 3000 {
			t.Errorf("expected default port 3000, got %d", port)
		}
	})

	t.Run("positional argument overrides default", func(t *testing.T) {
		tmpDir := t.TempDir()
		prevDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer os.Chdir(prevDir)

		cmd := newTunnelCmd()
		port, err := resolvePort(cmd, []string{"8080"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if port != 8080 {
			t.Errorf("expected port 8080, got %d", port)
		}
	})

	t.Run("flag overrides positional argument", func(t *testing.T) {
		tmpDir := t.TempDir()
		prevDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer os.Chdir(prevDir)

		cmd := newTunnelCmd()
		_ = cmd.Flags().Set("port", "9000")

		port, err := resolvePort(cmd, []string{"8080"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if port != 9000 {
			t.Errorf("expected flag port 9000 to override arg, got %d", port)
		}
	})

	t.Run("invalid non-numeric positional arg returns error", func(t *testing.T) {
		cmd := newTunnelCmd()
		_, err := resolvePort(cmd, []string{"invalid"})
		if err == nil {
			t.Error("expected error for non-numeric port, got nil")
		}
	})

	t.Run("out of range port returns error", func(t *testing.T) {
		cmd := newTunnelCmd()
		_, err := resolvePort(cmd, []string{"70000"})
		if err == nil {
			t.Error("expected error for out of range port, got nil")
		}
	})
}
