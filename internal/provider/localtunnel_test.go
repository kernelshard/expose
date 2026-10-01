package provider

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func Test_NewLocalTunnel(t *testing.T) {
	t.Run("with nil httpClient should use default", func(t *testing.T) {
		provider := NewLocalTunnel(nil)
		lt := provider.(*localTunnel)

		if lt.httpClient == nil {
			t.Fatal("expected default httpClient, got nil")
		}

		if lt.httpClient.Timeout != httpClientTimeout {
			t.Errorf("expected %v timeout, got %v", http.DefaultClient.Timeout, lt.httpClient.Timeout)
		}

		if lt.serverAPIEndpoint != localtunnelAPI {
			t.Errorf("expected endpoint %s, got %s", localtunnelAPI, lt.serverAPIEndpoint)
		}

		if cap(lt.connections) != clientMaxConn {
			t.Errorf("expected connections capacity %d, got %d", clientMaxConn, cap(lt.connections))
		}
	})

	t.Run("with custom httpClient should use it", func(t *testing.T) {
		customClient := &http.Client{Timeout: 5 * time.Second}

		provider := NewLocalTunnel(customClient)
		lt := provider.(*localTunnel)

		if lt.httpClient != customClient {
			t.Error("expected custom client to be used")
		}

		if lt.httpClient.Timeout != 5*time.Second {
			t.Errorf("expected 5s timeout, got %d", lt.httpClient.Timeout)
		}
	},
	)

}

// Test_requestTunnel tests the API call
func Test_requestTunnel(t *testing.T) {
	t.Run("successful API call", func(t *testing.T) {
		dummyRespID := "abc123"
		dummyRespURL := "https://abc123.example.com"
		dummyRespPort := 666666

		// mock handler
		mockHandler := func(w http.ResponseWriter, r *http.Request) {
			_, newQueryPresent := r.URL.Query()["new"]
			if r.URL.Path != "/" || !newQueryPresent {
				t.Error("expected /?new endpoint")
			}
			response := TunnelInfo{
				ID:      dummyRespID,
				URL:     dummyRespURL,
				Port:    dummyRespPort,
				MaxConn: clientMaxConn,
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		}

		// mock server
		server := httptest.NewServer(http.HandlerFunc(mockHandler))

		defer server.Close()

		lt := localTunnel{
			httpClient:        server.Client(),
			serverAPIEndpoint: server.URL,
		}

		ctx := context.Background()
		info, err := lt.requestTunnel(ctx)

		if err != nil {
			t.Fatalf("unexpected error:%v", err)
		}

		if info.ID != dummyRespID {
			t.Errorf("expected ID %s, got %s", dummyRespID, info.ID)
		}

		if info.URL != dummyRespURL {
			t.Errorf("expected URL %s, got %s", dummyRespURL, info.URL)
		}

		if info.Port != dummyRespPort {
			t.Errorf("expected Port %d, got %d", dummyRespPort, info.Port)
		}
	})

	t.Run("non-200 status code", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal server error"))
		}))

		defer server.Close()

		lt := localTunnel{httpClient: http.DefaultClient, serverAPIEndpoint: server.URL}

		ctx := context.Background()
		_, err := lt.requestTunnel(ctx)
		if err == nil {
			t.Fatalf("expected error for non-200 status")
		}

		if !strings.Contains(err.Error(), "500") {
			t.Errorf("expected error to maintain status 500, got %v", err)
		}
	})

	t.Run("invalid JSON response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("invalid json"))
		}))

		defer server.Close()

		lt := &localTunnel{
			httpClient:        http.DefaultClient,
			serverAPIEndpoint: server.URL,
		}

		ctx := context.Background()
		_, err := lt.requestTunnel(ctx)

		if err == nil {
			t.Fatalf("expected decode error for invalid JSON")
		}

		if !strings.Contains(err.Error(), "decode error") {
			t.Errorf("expected decode error, got %v", err)
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))

		defer server.Close()

		lt := &localTunnel{
			httpClient:        http.DefaultClient,
			serverAPIEndpoint: server.URL,
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // cancel immediately

		_, err := lt.requestTunnel(ctx)

		if err == nil {
			t.Fatal("expected error for cancelled context")
		}
	})
}

// TestLocalTunnel_Name
func TestLocalTunnel_Name(t *testing.T) {
	provider := NewLocalTunnel(nil)

	if provider.Name() != localTunnelProviderName {
		t.Errorf("expected name %s, got %s", localTunnelProviderName, provider.Name())
	}
}

// Test_IsConnected verifies connection state tracking
func TestLocalTunnel_IsConnected(t *testing.T) {
	lt := localTunnel{}
	if lt.IsConnected() {
		t.Errorf("new tunnel should not be connected")
	}

	lt.mu.Lock()
	lt.connected = true
	lt.mu.Unlock()

	if !lt.IsConnected() {
		t.Error("expected IsConnected to turn true")
	}

	lt.mu.Lock()
	lt.connected = false
	lt.mu.Unlock()

	if lt.IsConnected() {
		t.Errorf("expected IsConnected to return false after disconnect")
	}
}

// TestLocalTunnel_PublicURL verifies URL getter
func TestLocalTunnel_PublicURL(t *testing.T) {

	tests := []struct {
		name string
		url  string
	}{
		{"empty URL", ""},
		{"valid URL", "https://test.localtunnel.me"},
		{"custom domain", "https://abc123.loca.lt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lt := &localTunnel{publicURL: tt.url}
			got := lt.PublicURL()

			if got != tt.url {
				t.Errorf("expected URL %s, got %s", tt.url, got)
			}
		})
	}
}

func Test_closeAllConnections(t *testing.T) {
	// create mock connection s
	conn1Client, conn1Server := net.Pipe()
	conn2Client, conn2Server := net.Pipe()

	defer conn1Server.Close()
	defer conn2Server.Close()

	lt := &localTunnel{
		connections: []net.Conn{conn1Client, conn2Client},
	}

	lt.closeAllConnections()
	// make sure all closed
	if len(lt.connections) != 0 {
		t.Errorf("expected empty connections slice, got %d connections, ", len(lt.connections))
	}

	// verify connections are actually closed,
	_, err := conn1Client.Write([]byte("test"))
	if err == nil {
		t.Error("expected error writing to closed connection")
	}
	_, err = conn2Client.Write([]byte("test"))
	if err == nil {
		t.Error("Expected error writing to closed connection")
	}

}

func TestLocalTunnel_Close(t *testing.T) {
	ctx, canelFunc := context.WithCancel(context.Background())

	// create mock connection
	clientConn, serverConn := net.Pipe()
	defer serverConn.Close()

	lt := &localTunnel{
		connected:   true,
		publicURL:   "https://test.example.com",
		connections: []net.Conn{clientConn},
		cancel:      canelFunc,
	}

	err := lt.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if lt.IsConnected() {
		t.Error("expected connected to be false after Close")
	}
	if len(lt.connections) != 0 {
		t.Errorf("expected connections to be cleared, got %d", len(lt.connections))
	}

	// verify ctx was canceled
	select {
	case <-ctx.Done():
	// expected
	default:
		t.Error("expected context to be cancelled")
	}

}

func TestLocalTunnel_ProxyRequest(t *testing.T) {
	// 1. Start a dummy local app
	localApp := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello localtunnel"))
	}))
	defer localApp.Close()

	localPort := localApp.Listener.Addr().(*net.TCPAddr).Port

	// 2. Create in-memory pipe simulating the tunnel
	tunnelClient, tunnelServer := net.Pipe()
	defer tunnelClient.Close()

	lt := &localTunnel{localPort: localPort}

	// 3. Run proxyRequest in the background
	go lt.proxyRequest(tunnelServer)

	// 4. Send HTTP request
	req := "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n"
	if _, err := tunnelClient.Write([]byte(req)); err != nil {
		t.Fatalf("failed to write request: %v", err)
	}

	// 5. Read response
	buf := make([]byte, 1024)
	n, err := tunnelClient.Read(buf)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}
	if !strings.Contains(string(buf[:n]), "hello localtunnel") {
		t.Errorf("expected response containing 'hello localtunnel', got: %s", string(buf[:n]))
	}

	// 6. Test failure when local port is closed
	dummyClient, dummyServer := net.Pipe()
	defer dummyClient.Close()
	ltClosed := &localTunnel{localPort: 0}
	if err := ltClosed.proxyRequest(dummyServer); err == nil {
		t.Errorf("expected error when localPort is closed, got nil")
	}
}

func TestLocalTunnel_DialTunnel(t *testing.T) {
	// 1. Success: Start a test listener
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer ln.Close()

	port := ln.Addr().(*net.TCPAddr).Port
	lt := &localTunnel{
		tunnelHost: "127.0.0.1",
		tunnelPort: port,
	}

	conn, err := lt.dialTunnel()
	if err != nil {
		t.Fatalf("dialTunnel failed: %v", err)
	}
	conn.Close()

	// 2. Failure: Dialing an unreachable port (port 0)
	ltFail := &localTunnel{
		tunnelHost: "127.0.0.1",
		tunnelPort: 0,
	}
	_, err = ltFail.dialTunnel()
	if err == nil {
		t.Errorf("expected error dialing port 0, got nil")
	}
}

func TestLocalTunnel_OpenConnections(t *testing.T) {
	// 1. Success: Start a test TCP server that accepts connections
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer ln.Close()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			defer conn.Close()
		}
	}()

	ctx := t.Context()

	port := ln.Addr().(*net.TCPAddr).Port
	lt := &localTunnel{
		tunnelHost:     "127.0.0.1",
		tunnelPort:     port,
		maxConnections: 2,
		ctx:            ctx,
	}

	if err := lt.openConnections(); err != nil {
		t.Fatalf("openConnections failed: %v", err)
	}

	if len(lt.connections) != 2 {
		t.Errorf("expected 2 connections in pool, got %d", len(lt.connections))
	}
	lt.closeAllConnections()

	// 2. Failure: When tunnel port is unreachable
	ltFail := &localTunnel{
		tunnelHost:     "127.0.0.1",
		tunnelPort:     0,
		maxConnections: 1,
	}
	if err := ltFail.openConnections(); err == nil {
		t.Errorf("expected error when dialing unreachable port, got nil")
	}
}

func TestLocalTunnel_Connect_RequestTunnelError(t *testing.T) {
	// 1. Mock server that returns a 500 error
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	}))
	defer mockServer.Close()

	provider := NewLocalTunnel(mockServer.Client())
	lt := provider.(*localTunnel)
	lt.serverAPIEndpoint = mockServer.URL

	// 2. Calling Connect should fail with an error from requestTunnel
	_, err := lt.Connect(t.Context(), 3000)
	if err == nil || !strings.Contains(err.Error(), "failed to request tunnel") {
		t.Errorf("expected 'failed to request tunnel' error, got: %v", err)
	}
}

func TestLocalTunnel_Connect_Success(t *testing.T) {
	// 1. Mock TCP server for the data connections
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer ln.Close()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			defer conn.Close()
		}
	}()

	tcpPort := ln.Addr().(*net.TCPAddr).Port

	// 2. Mock HTTP server for requestTunnel
	mockHTTPServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := TunnelInfo{
			ID:      "test-id",
			URL:     "https://test.localtunnel.me",
			Port:    tcpPort,
			MaxConn: 2,
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockHTTPServer.Close()

	// 3. Connect using our local test endpoints
	provider := NewLocalTunnel(mockHTTPServer.Client())
	lt := provider.(*localTunnel)
	lt.serverAPIEndpoint = mockHTTPServer.URL
	lt.tunnelHost = "127.0.0.1"

	url, err := lt.Connect(t.Context(), 3000)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	if url != "https://test.localtunnel.me" {
		t.Errorf("expected URL https://test.localtunnel.me, got: %s", url)
	}
	if !lt.IsConnected() {
		t.Errorf("expected IsConnected to be true")
	}

	_ = lt.Close()
}
