package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
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
			t.Errorf("Server error: %v", err)
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

	// First register a subdomain so a ClientConnection exists
	regClient, regServer := net.Pipe()
	defer regClient.Close()
	defer regServer.Close()

	sub, err := srv.registerSubdomain(regServer, "alice")
	if err != nil {
		t.Fatalf("failed to register subdomain: %v", err)
	}

	val, ok := srv.tunnels.Load(sub)
	if !ok {
		t.Fatalf("subdomain not found in tunnels map")
	}
	client := val.(ClientConnection)

	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()

	go srv.handleControlConnection(serverConn)

	// send data connection request targeting the registered subdomain
	req := protocol.TunnelRequest{
		Type:      protocol.TypeData,
		Subdomain: sub,
	}
	if err := json.NewEncoder(clientConn).Encode(req); err != nil {
		t.Fatalf("failed to encode request: %v", err)
	}

	// Should be pushed to client.dataConns
	select {
	case conn := <-client.dataConns:
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

	// Send garbage bytes that cannot be parsed as JSON
	_, _ = clientConn.Write([]byte("invalid-json\n"))

	// Server should close the connection
	buf := make([]byte, 10)
	_, err := clientConn.Read(buf)
	if err == nil {
		t.Errorf("expected connection to be closed on invalid JSON")
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

	// Register subdomain
	sub, err := srv.registerSubdomain(c2, "mytest")
	if err != nil {
		t.Fatalf("failed to register subdomain: %v", err)
	}
	val, _ := srv.tunnels.Load(sub)
	client := val.(ClientConnection)

	// Place a data connection in the client's queue so ServeHTTP doesn't block
	client.dataConns <- c1

	// httptest.ResponseRecorder does not support hijacking, so this triggers the error branch
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "mytest.localhost"

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestServer_Success(t *testing.T) {
	srv := NewServer("localhost", 0, 0)

	// create a mock tunnel to simulate the laptop
	tunnelClientConn, tunnelServerConn := net.Pipe()
	defer tunnelClientConn.Close()

	// Register subdomain
	sub, err := srv.registerSubdomain(tunnelClientConn, "alice")
	if err != nil {
		t.Fatalf("failed to register subdomain: %v", err)
	}
	val, _ := srv.tunnels.Load(sub)
	client := val.(ClientConnection)

	client.dataConns <- tunnelServerConn

	ts := httptest.NewServer(srv)
	defer ts.Close()

	// simulate laptop receiving the request and sending back 200 ok
	go func() {
		buf := make([]byte, 1024)
		n, _ := tunnelClientConn.Read(buf)

		if n > 0 {
			resp := "HTTP/1.1 200 OK\r\nContent-Length: 12\r\n\r\nHello World!"
			_, _ = tunnelClientConn.Write([]byte(resp))
		}
	}()

	// make a real HTTP request to the public server with Host header
	req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Host = "alice.localhost"

	resp, err := http.DefaultClient.Do(req)
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

func TestServer_IsClosed(t *testing.T) {
	srv := NewServer("localhost", 0, 0)

	if srv.isClosed(nil) {
		t.Errorf("expected isClosed(nil) to be false")
	}

	if srv.isClosed(errors.New("unexpected network timeout")) {
		t.Errorf("expected isClosed(network error) to be false")
	}

	if !srv.isClosed(net.ErrClosed) {
		t.Errorf("expected isClosed(net.ErrClosed) to be true")
	}

	if !srv.isClosed(fmt.Errorf("test: %w", net.ErrClosed)) {
		t.Errorf("expected isClosed(wrapped net.ErrClosed) to be true")
	}
}

func TestSever_HandleControl_SubdomainAlreadyTaken(t *testing.T) {
	srv := NewServer("tunnel.example.com", 7890, 8080)

	// Client 1 claims "my-testapp"
	c1, s1 := net.Pipe()
	defer c1.Close()

	go srv.handleControlConnection(s1)

	req := protocol.TunnelRequest{
		Type: protocol.TypeControl, Subdomain: "my-testapp",
	}
	if err := json.NewEncoder(c1).Encode(req); err != nil {
		t.Fatalf("c1 encode error: %v", err)
	}

	var resp1 protocol.TunnelResponse
	if err := json.NewDecoder(c1).Decode(&resp1); err != nil {
		t.Fatalf("c1 decode error: %v", err)
	}

	// Client 2 claims the same subdomain and expects an error
	c2, s2 := net.Pipe()
	defer c2.Close()

	go srv.handleControlConnection(s2)

	if err := json.NewEncoder(c2).Encode(req); err != nil {
		t.Fatalf("c2 encode error: %v", err)
	}

	var resp2 protocol.TunnelResponse
	if err := json.NewDecoder(c2).Decode(&resp2); err != nil {
		t.Fatalf("c2 decode error: %v", err)
	}

	if resp2.Error == "" {
		t.Fatalf("expected collision error for client 2, got empty error")
	}

	if resp2.PublicURL != "" {
		t.Fatalf("expected no public URL for client 2 collision, got %q", resp2.PublicURL)
	}
}

func TestExtractSubdomain(t *testing.T) {
	tests := []struct {
		name       string
		host       string
		baseDomain string
		want       string
	}{
		{"standard subdomain", "alice.localhost", "localhost", "alice"},
		{"with port", "alice.tunnel.com:8080", "tunnel.com", "alice"},
		{"uppercase normalization", "ALICE.TUNNEL.COM", "tunnel.com", "alice"},
		{"nested subdomain", "staging.alice.tunnel.com", "tunnel.com", "staging.alice"},
		{"apex domain rejected", "tunnel.com", "tunnel.com", ""},
		{"apex with port rejected", "tunnel.com:8080", "tunnel.com", ""},
		{"raw IPv4 rejected", "127.0.0.1", "localhost", ""},
		{"raw IPv4 with port rejected", "192.168.1.1:8080", "tunnel.com", ""},
		{"unrelated domain rejected", "evil.com", "tunnel.com", ""},
		{"empty base domain returns empty", "alice.localhost", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractSubdomain(tt.host, tt.baseDomain)
			if got != tt.want {
				t.Errorf("extractSubdomain(%q, %q) = %q, want %q", tt.host, tt.baseDomain, got, tt.want)
			}
		})
	}
}

func TestServer_ServeHTTP_InvalidSubdomain(t *testing.T) {
	srv := NewServer("localhost", 0, 0)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "localhost" // apex domain, no subdomain

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestServer_ServeHTTP_TunnelNotFound(t *testing.T) {
	srv := NewServer("localhost", 0, 0)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "unknown.localhost"

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusBadGateway)
	}
}

func TestServer_ServeHTTP_ContextCanceled(t *testing.T) {
	srv := NewServer("localhost", 0, 0)
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	sub, err := srv.registerSubdomain(c1, "alice")
	if err != nil {
		t.Fatalf("failed to register subdomain: %v", err)
	}
	_ = sub

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	req.Host = "alice.localhost"

	srv.ServeHTTP(rec, req)
	// Should return cleanly without hanging
}

func TestServer_MultiTenant_Routing(t *testing.T) {
	srv := NewServer("localhost", 0, 0)

	// Setup Alice
	aliceClient, aliceServer := net.Pipe()
	defer aliceClient.Close()
	_, err := srv.registerSubdomain(aliceClient, "alice")
	if err != nil {
		t.Fatalf("failed to register alice: %v", err)
	}
	valA, _ := srv.tunnels.Load("alice")
	alice := valA.(ClientConnection)
	alice.dataConns <- aliceServer

	// Setup Bob
	bobClient, bobServer := net.Pipe()
	defer bobClient.Close()
	_, err = srv.registerSubdomain(bobClient, "bob")
	if err != nil {
		t.Fatalf("failed to register bob: %v", err)
	}
	valB, _ := srv.tunnels.Load("bob")
	bob := valB.(ClientConnection)
	bob.dataConns <- bobServer

	ts := httptest.NewServer(srv)
	defer ts.Close()

	// Simulate Alice's laptop responding
	go func() {
		defer aliceClient.Close()
		buf := make([]byte, 1024)
		n, _ := aliceClient.Read(buf)
		if n > 0 {
			resp := "HTTP/1.1 200 OK\r\nContent-Length: 12\r\nConnection: close\r\n\r\nHello Alice!"
			_, _ = aliceClient.Write([]byte(resp))
		}
	}()

	// Simulate Bob's laptop responding
	go func() {
		defer bobClient.Close()
		buf := make([]byte, 1024)
		n, _ := bobClient.Read(buf)
		if n > 0 {
			resp := "HTTP/1.1 200 OK\r\nContent-Length: 10\r\nConnection: close\r\n\r\nHello Bob!"
			_, _ = bobClient.Write([]byte(resp))
		}
	}()

	httpClient := &http.Client{
		Transport: &http.Transport{DisableKeepAlives: true},
	}

	// Request for Alice
	reqAlice, _ := http.NewRequest(http.MethodGet, ts.URL, nil)
	reqAlice.Host = "alice.localhost"
	respAlice, err := httpClient.Do(reqAlice)
	if err != nil {
		t.Fatalf("failed to request alice: %v", err)
	}
	defer respAlice.Body.Close()
	bodyAlice, _ := io.ReadAll(respAlice.Body)
	if string(bodyAlice) != "Hello Alice!" {
		t.Errorf("expected %q, got %q", "Hello Alice!", string(bodyAlice))
	}

	// Request for Bob
	reqBob, _ := http.NewRequest(http.MethodGet, ts.URL, nil)
	reqBob.Host = "bob.localhost"
	respBob, err := httpClient.Do(reqBob)
	if err != nil {
		t.Fatalf("failed to request bob: %v", err)
	}
	defer respBob.Body.Close()
	bodyBob, _ := io.ReadAll(respBob.Body)
	if string(bodyBob) != "Hello Bob!" {
		t.Errorf("expected %q, got %q", "Hello Bob!", string(bodyBob))
	}
}

func Test_HandleControl_KeepAliveAndDisconnect(t *testing.T) {
	srv := NewServer("tunnel.localhost.com", 7890, 8080)
	clientConn, serverConn := net.Pipe()
	go srv.handleControlConnection(serverConn)

	// 1. Send registration request
	req := protocol.TunnelRequest{Type: protocol.TypeControl, Subdomain: "example"}
	_ = json.NewEncoder(clientConn).Encode(req)

	// 2. Read the server's response
	var resp protocol.TunnelResponse
	_ = json.NewDecoder(clientConn).Decode(&resp)

	// 3. Check is the subdomain actually stored in srv.tunnels
	if _, ok := srv.tunnels.Load("example"); !ok {
		t.Fatal("expected subdomain to be registered in map")
	}

	// client sends ping to server
	for range 3 {
		if _, err := clientConn.Write([]byte("PING\n")); err != nil {
			t.Fatalf("failed to send ping: %v", err)
		}
	}

	// tunnel must still be alive in srv.tunnels
	if _, ok := srv.tunnels.Load("example"); !ok {
		t.Fatal("tunnel was unexpectedly removed")
	}

	// client closed connection
	clientConn.Close()

	// give the goroutine time to detect the disconnection and remove the tunnel
	time.Sleep(50 * time.Millisecond)
	// tunnel must be removed now
	if _, ok := srv.tunnels.Load("example"); ok {
		t.Fatal("expected subdomain to be deleted after the client disconnects")
	}
}

func Test_HandleControl_HeartbeatTimeout(t *testing.T) {
	srv := NewServer("tunnel.localhost.com", 7890, 8080)
	srv.SetControlTimeout(50 * time.Millisecond) // Evict after 50ms of silence

	clientConn, serverConn := net.Pipe()
	go srv.handleControlConnection(serverConn)

	// 1. Register tunnel
	req := protocol.TunnelRequest{Type: protocol.TypeControl, Subdomain: "ghost-app"}
	_ = json.NewEncoder(clientConn).Encode(req)

	var resp protocol.TunnelResponse
	_ = json.NewDecoder(clientConn).Decode(&resp)

	// 2. Verify registered
	if _, ok := srv.tunnels.Load("ghost-app"); !ok {
		t.Fatal("expected ghost-app to be registered")
	}

	// 3. Client goes completely silent! (simulating lid closed/battery dead)
	// We wait past the 50ms deadline:
	time.Sleep(100 * time.Millisecond)

	// 4. Verify the server automatically evicted the zombie tunnel!
	if _, ok := srv.tunnels.Load("ghost-app"); ok {
		t.Fatal("expected ghost-app to be evicted after heartbeat timeout")
	}

}

func Test_HandleControl_HandshakeTimeout(t *testing.T) {
	srv := NewServer("localhost", 7890, 8080)
	srv.SetHandshakeTimeout(50 * time.Millisecond)
	clientConn, serverConn := net.Pipe()
	go srv.handleControlConnection(serverConn)

	// Send raw unclosed '{' without trailing quote or newline:
	_, _ = clientConn.Write([]byte("{"))
	time.Sleep(100 * time.Millisecond)

	buf := make([]byte, 10)
	_, err := clientConn.Read(buf)

	// connection has been closed due to the sleep > handshake timeout
	if err != io.EOF {
		t.Fatal("expected connection to be closed after handshake timeout")
	}
}

func Test_HandleControl_OversizePayload(t *testing.T) {
	srv := NewServer("localhost", 7890, 8080)

	clientConn, serverConn := net.Pipe()
	go srv.handleControlConnection(serverConn)

	// Send an oversized payload (e.g., 5KB of JSON)
	sub := strings.Repeat("a", 5000)
	oversizedJSON := fmt.Sprintf(`{"subdomain": "%s"}`, sub)

	// net.Pipe blocks until all 5000 bytes are read, but LimitReader stops at 4096.
	// so we need to use a goroutine to send the oversized payload
	go func() {
		_, _ = clientConn.Write([]byte(oversizedJSON))
	}()

	// Wait for the server to process the oversized payload
	time.Sleep(100 * time.Millisecond)

	// Verify the connection is closed
	buf := make([]byte, 10)
	_, err := clientConn.Read(buf)

	if err != io.EOF {
		t.Errorf("expected connection to be closed after oversize payload, got error: %v", err)
	}

	// Verify the subdomain was not registered
	if _, ok := srv.tunnels.Load(sub); ok {
		t.Error("expected subdomain not to be registered for oversize payload")
	}
}

func TestServer_HandleControl_RejectInvalidSubdomain(t *testing.T) {
	srv := NewServer("tunnel.example.com", 7890, 8080)
	clientConn, serverConn := net.Pipe()
	go srv.handleControlConnection(serverConn)

	// Attempt to register reserved subdomain "admin"
	req := protocol.TunnelRequest{Type: protocol.TypeControl, Subdomain: "admin"}
	_ = json.NewEncoder(clientConn).Encode(req)

	var resp protocol.TunnelResponse
	_ = json.NewDecoder(clientConn).Decode(&resp)

	if resp.Error == "" {
		t.Fatal("expected error registering reserved subdomain 'admin', got empty")
	}
	if resp.PublicURL != "" {
		t.Fatalf("expected empty public URL on error, got %q", resp.PublicURL)
	}
}

func TestServer_ServeHTTP_SymmetricTeardown_VisitorDisconnect(t *testing.T) {
	srv := NewServer("localhost", 0, 0)

	// Actor 1: The Mock Laptop Tunnel
	mockLaptop, serverDataConn := net.Pipe()
	defer mockLaptop.Close()

	// Register "alice"
	sub, err := srv.registerSubdomain(mockLaptop, "alice")
	if err != nil {
		t.Fatalf("failed to register subdomain: %v", err)
	}
	val, _ := srv.tunnels.Load(sub)
	client := val.(ClientConnection)
	client.dataConns <- serverDataConn

	ts := httptest.NewServer(srv)
	defer ts.Close()

	// Actor 2: The Public Web Visitor (direct TCP connection to test server)
	mockVisitor, err := net.Dial("tcp", ts.Listener.Addr().String())
	if err != nil {
		t.Fatalf("failed to dial test server: %v", err)
	}

	// 1. mockVisitor sends the HTTP request headers
	req := "GET / HTTP/1.1\r\nHost: alice.localhost\r\n\r\n"
	if _, err := mockVisitor.Write([]byte(req)); err != nil {
		t.Fatalf("failed to write request: %v", err)
	}

	// 2. mockLaptop receives the forwarded request
	buf := make([]byte, 1024)
	n, err := mockLaptop.Read(buf)
	if err != nil || n == 0 {
		t.Fatalf("tunnel failed to read forwarded request: %v", err)
	}

	// 3. mockVisitor abruptly hangs up / disconnects
	_ = mockVisitor.Close()

	// 4. With symmetric teardown, the server MUST close serverDataConn so mockLaptop unblocks with EOF
	readDone := make(chan error, 1)
	go func() {
		_, err := mockLaptop.Read(buf)
		readDone <- err
	}()

	select {
	case err := <-readDone:
		if err != io.EOF && !errors.Is(err, net.ErrClosed) {
			t.Errorf("expected EOF on laptop socket after visitor disconnect, got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("hung: laptop socket was not closed when visitor disconnected (socket leak!)")
	}
}
