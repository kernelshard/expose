package cli

import (
	"context"
	"errors"
	"net"
	"os"
	"testing"
	"time"

	"github.com/kernelshard/expose/internal/tunnel"
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

	subdomainFlag := cmd.Flags().Lookup("subdomain")
	if subdomainFlag == nil {
		t.Error("subdomain flag not defined")
	}
	if subdomainFlag.Shorthand != "d" {
		t.Errorf("expected shorthand 'd', got %s", subdomainFlag.Shorthand)
	}

	tokenFlag := cmd.Flags().Lookup("token")
	if tokenFlag == nil {
		t.Error("token flag not defined")
	}
	if tokenFlag.Shorthand != "t" {
		t.Errorf("expected shorthand 't', got %s", tokenFlag.Shorthand)
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

func TestIsPortListening(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer ln.Close()

	port := ln.Addr().(*net.TCPAddr).Port
	if !isPortListening(port) {
		t.Errorf("expected port %d to be listening", port)
	}

	_ = ln.Close()
	if isPortListening(port) {
		t.Errorf("expected closed port %d to not be listening", port)
	}
}

type mockProvider struct {
	name      string
	publicURL string
	connectFn func(ctx context.Context, localPort int) (string, error)
	closeFn   func() error
}

func (m *mockProvider) Connect(ctx context.Context, localPort int) (string, error) {
	if m.connectFn != nil {
		return m.connectFn(ctx, localPort)
	}
	return m.publicURL, nil
}

func (m *mockProvider) Close() error {
	if m.closeFn != nil {
		return m.closeFn()
	}
	return nil
}

func (m *mockProvider) IsConnected() bool { return true }
func (m *mockProvider) PublicURL() string { return m.publicURL }
func (m *mockProvider) Name() string      { return m.name }

func TestRunTunnelWithContext_Success(t *testing.T) {
	mock := &mockProvider{
		name:      "mock",
		publicURL: "https://test.expose.dev",
	}
	svc := tunnel.NewService(mock)

	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		errCh <- runTunnelWithContext(ctx, 3000, svc)
	}()

	select {
	case <-svc.Ready():
		cancel()
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for service to be ready")
	}

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("expected nil error on clean shutdown, got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for runTunnelWithContext to finish")
	}
}

func TestRunTunnelWithContext_StartError(t *testing.T) {
	mock := &mockProvider{
		name: "mock",
		connectFn: func(ctx context.Context, localPort int) (string, error) {
			return "", errors.New("connection failed")
		},
	}
	svc := tunnel.NewService(mock)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := runTunnelWithContext(ctx, 3000, svc)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
