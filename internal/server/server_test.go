package server

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestServerLifeCycle(t *testing.T) {
	// create a server with random port (0)
	srv := NewServer("localhost", 0, 0)

	// start in the background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- srv.Start(ctx)
	}()

	// wait for listener to bind
	time.Sleep(600 * time.Millisecond)

	// Verify listeners exist
	if srv.controlListener == nil {
		t.Fatalf("Control listener not started")
	}

	// Connect to control plane
	controlAddr := srv.controlListener.Addr().String()
	conn, err := net.Dial("tcp", controlAddr)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// stop
	cancel()

	// wait for server to stop
	select {
	case err := <-errChan:
		if err != nil {
			t.Errorf("Server errror: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for stop!")
	}

}
