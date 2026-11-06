package cli

import (
	"testing"
)

func TestServerCmd(t *testing.T) {
	cmd := newServerCmd()

	if cmd == nil {
		t.Fatal("newServerCmd returned nil")
	}

	if cmd.Name() != "server" {
		t.Errorf("expected command name 'server', got '%s'", cmd.Name())
	}

	flags := []struct {
		name      string
		shorthand string
	}{
		{"control-port", "c"},
		{"public-port", "p"},
		{"domain", "d"},
		{"token", "t"},
	}

	for _, f := range flags {
		flag := cmd.Flags().Lookup(f.name)
		if flag == nil {
			t.Errorf("flag %q not defined", f.name)
			continue
		}
		if flag.Shorthand != f.shorthand {
			t.Errorf("flag %q: expected shorthand %q, got %q", f.name, f.shorthand, flag.Shorthand)
		}
	}
}
