package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kernelshard/expose/internal/server/protocol"
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
	select {
	case <-srv.Ready():
		// ready to accept connections
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for server to be ready")
	}

	// Connect to control plane
	controlAddr := srv.ControlAddr()
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

func TestServer_ControlAddr_NotStarted(t *testing.T) {
	srv := NewServer("localhost", 7890, 8080)
	if addr := srv.ControlAddr(); addr != "" {
		t.Errorf("expected empty string when not started, got: %s", addr)
	}
}

func TestServer_HandleControl_RegisterCustomSubdomain(t *testing.T) {
	srv := NewServer("tunnel.example.com", 7890, 8080)

	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()

	go srv.handleControlConnection(serverConn)

	// send register request with custom subdomain
	req := protocol.TunnelRequest{
		Type:      protocol.TypeControl,
		Subdomain: "my-testapp",
	}
	if err := json.NewEncoder(clientConn).Encode(req); err != nil {
		t.Fatalf("failed to encode request %v", err)
	}

	// read response
	var resp protocol.TunnelResponse
	if err := json.NewDecoder(clientConn).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response %v", err)
	}

	if resp.Subdomain != "my-testapp" {
		t.Errorf("expected subdomain %q, got %q", "my-testapp", resp.Subdomain)
	}

	if resp.PublicURL != "http://my-testapp.tunnel.example.com:8080" {
		t.Errorf("expected public URL %q, got %q", "http://my-testapp.tunnel.example.com:8080", resp.PublicURL)
	}
}

func TestServer_HandleControl_DataConnection(t *testing.T) {
	srv := NewServer("localhost", 7890, 8080)
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()

	go srv.handleControlConnection(serverConn)

	// send data connection request
	req := protocol.TunnelRequest{
		Type: protocol.TypeData,
	}
	if err := json.NewEncoder(clientConn).Encode(req); err != nil {
		t.Fatalf("failed to encode request: %v", err)
	}

	// Should be pushed to srv.dataConns
	select {
	case conn := <-srv.dataConns:
		if conn == nil {
			t.Fatalf("expected non-nil data connection")
		}
		defer conn.Close()
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for data connection on channel")
	}
}

func TestServer_HandleControl_InvalidJSON(t *testing.T) {
	srv := NewServer("localhost", 7890, 8080)
	clientConn, serverConn := net.Pipe()

	go srv.handleControlConnection(serverConn)

	// Send garbage bytes that canbot be parsed as JSON
	_, _ = clientConn.Write([]byte("invalid-json\n"))

	// Server should close the conenction
	buf := make([]byte, 10)
	_, err := clientConn.Read(buf)
	if err == nil {
		t.Errorf("expected connection to be cosed on invalid JSON")
	}

	// Ignore EOF and network errors
	if !errors.Is(err, io.EOF) {
		t.Errorf("expected EOF error, got: %v", err)
	}
}

func TestServer_HandleControl_EmptySubdomainGeneratesRandom(t *testing.T) {
	srv := NewServer("tunnel.example.com", 7890, 8080)
	clientConn, serverConn := net.Pipe()

	go srv.handleControlConnection(serverConn)

	// send req without specifying a subdomain
	req := protocol.TunnelRequest{
		Type: protocol.TypeControl,
	}

	if err := json.NewEncoder(clientConn).Encode(req); err != nil {
		t.Fatalf("failed to encode request: %v", err)
	}

	var resp protocol.TunnelResponse
	if err := json.NewDecoder(clientConn).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Subdomain == "" {
		t.Errorf("expected server to generate a random subdomain, got empty string")
	}
}

// Tests that ServeHTTP safely returns 500 when ResponseWriter doesn't support TCP hijacking (e.g. test recorders).
func TestServer_ServeHTTP_NonHijacker(t *testing.T) {
	srv := NewServer("localhost", 0, 0)
	c1, c2 := net.Pipe()

	defer c1.Close()
	defer c2.Close()

	// Place a data connection in the queue so ServeHTTP doesn't block
	srv.dataConns <- c1

	// httptest.ResponseRecorder does not support hijacking, so this triggers the error branch
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestServer_Success(t *testing.T) {
	srv := NewServer("localhost", 0, 0)

	// create a mock tunnel to simulate the laptop
	tunnelClientCOnn, tunnelServerConn := net.Pipe()
	defer tunnelClientCOnn.Close()

	srv.dataConns <- tunnelServerConn

	ts := httptest.NewServer(srv)
	defer ts.Close()

	// simulate laptop receiving the request and sending back 200 ok
	go func() {
		buf := make([]byte, 1024)
		n, _ := tunnelClientCOnn.Read(buf)

		if n > 0 {
			resp := "HTTP/1.1 200 OK\r\nContent-Length: 12\r\n\r\nHello World!"
			_, _ = tunnelClientCOnn.Write([]byte(resp))
		}
	}()

	// make a real HTTP request to the public server
	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("failed to make http request: %v", err)
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "Hello World!" {
		t.Errorf("expected body %q, got: %q", "Hello World!", string(body))
	}

}

func TestServer_Start_PortAlreadyInUse(t *testing.T) {
	// 1. Pre-occupy a port
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("failed to bind port: %v", err)
	}
	defer ln.Close()
	busyPort := ln.Addr().(*net.TCPAddr).Port
	// 2. Try to start expose server on the exact same port
	srv := NewServer("localhost", busyPort, 0)
	err = srv.Start(context.Background())
	if err == nil {
		t.Errorf("expected error when port is already in use, got nil")
	}
}
