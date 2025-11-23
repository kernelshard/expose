package provider

import (
	"context"
	"errors"
	"os/exec"
	"testing"
	"time"
)

// TestCloudflare_Connect tests the Connect method of Cloudflare provider
func TestCloudflare_Connect(t *testing.T) {
	cf := NewCloudFlare()

	// Mock RequestTunnel
	cf.RequestTunnel = func(ctx context.Context, port int, timeout time.Duration) (string, *exec.Cmd, error) {
		return "https://test-tunnel.trycloudflare.com", nil, nil
	}

	url, err := cf.Connect(context.Background(), 3000)
	if err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}

	if url != "https://test-tunnel.trycloudflare.com" {
		t.Errorf("got %s, want test URL", url)
	}

	if cf.PublicURL() != url {
		t.Errorf("PublicURL() = %s, want %s", cf.PublicURL(), url)
	}
}

// TestCloudflare_ConnectError tests the Connect method when RequestTunnel returns an error
func TestCloudflare_ConnectError(t *testing.T) {
	cf := NewCloudFlare()

	cf.RequestTunnel = func(ctx context.Context, port int, timeout time.Duration) (string, *exec.Cmd, error) {
		return "", nil, errors.New("tunnel creation failed")
	}

	_, err := cf.Connect(context.Background(), 3000)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if cf.PublicURL() != "" {
		t.Errorf("PublicURL() should be empty, got %s", cf.PublicURL())
	}
}

// TestCloudflare_Name tests the Name method of Cloudflare provider
func TestCloudflare_Name(t *testing.T) {
	cf := NewCloudFlare()
	if got := cf.Name(); got != "Cloudflare" {
		t.Errorf("Name() = %s, want Cloudflare", got)
	}
}

// TestCloudflare_CloseBeforeConnect tests that Close can be called before Connect without error
func TestCloudflare_CloseBeforeConnect(t *testing.T) {
	cf := NewCloudFlare()
	if err := cf.Close(); err != nil {
		t.Errorf("Close() before Connect error: %v", err)
	}
}

// TestCloudflare_ConnectTimeout tests the Connect method with a context timeout
func TestCloudflare_ConnectTimeout(t *testing.T) {
	cf := NewCloudFlare()
	// Mock RequestTunnel to simulate delay
	cf.RequestTunnel = func(ctx context.Context, port int, timeout time.Duration) (string, *exec.Cmd, error) {
		<-ctx.Done()
		return "", nil, ctx.Err()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := cf.Connect(ctx, 3000)
	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}
}
