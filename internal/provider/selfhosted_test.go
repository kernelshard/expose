package provider

import (
	"context"
	"testing"
	"time"

	"github.com/kernelshard/expose/internal/server"
)

func TestSelfHosted_Connect(t *testing.T) {
	// 1. start a real server locally
	srv := server.NewServer("localhost", 0, 0)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := srv.Start(ctx); err != nil {
			t.Errorf("server error: %v", err)
		}
	}()

	time.Sleep(200 * time.Millisecond) // wait for listeners

	// 2. Get the control address from the server
	controlAddr := srv.ControlAddr()

	// 3. Connect via selfhosted provider
	provider := NewSelfHosted(controlAddr)
	url, errr := provider.Connect(ctx, 3000)
	if errr != nil {
		t.Errorf("Connect failed: %v", errr)
	}

	// 4. Verify the connection
	if url == "" {
		t.Fatal("expected public URL, got empty string")
	}
	if !provider.IsConnected() {
		t.Fatal("expected IsConnected to be true")
	}

	// 5. Close the connection
	if err := provider.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}
