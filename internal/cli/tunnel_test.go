package cli

import (
	"testing"
)

func TestTunnelCmd(t *testing.T) {
	cmd := newTunnelCmd()

	if cmd == nil {
		t.Fatal("newTunnelCmd returned nil")
	}

	if cmd.Use != "tunnel" {
		t.Errorf("expected Use 'tunnel', got '%s'", cmd.Use)
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
