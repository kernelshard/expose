package provider

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/kernelshard/expose/internal/server"
	"github.com/kernelshard/expose/internal/server/protocol"
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
	<-srv.Ready()
	// 2. Get the control address from the server
	controlAddr := srv.ControlAddr()

	// 3. Connect via selfhosted provider
	provider := NewSelfHosted(controlAddr)
	url, err := provider.Connect(ctx, 3000)
	if err != nil {
		t.Errorf("Connect failed: %v", err)
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

func TestSelfHosted_GettersAndClose(t *testing.T) {
	p := NewSelfHosted("localhost:7890")

	if p.Name() != "selfhosted" {
		t.Errorf("expected provider name 'selfhosted', got '%s'", p.Name())
	}

	if p.PublicURL() != "" {
		t.Errorf("expected empty public URL before connect, got '%s'", p.PublicURL())

	}

	if p.IsConnected() {
		t.Errorf("expected IsConnected() to be false before connect, got true")
	}

	// calling Close on an unstarted provider should be safe
	if err := p.Close(); err != nil {
		t.Errorf("expected Close() to return nil, got: %v", err)
	}

}

func TestSelfHosted_Connect_Unreachable(t *testing.T) {
	provider := NewSelfHosted("localhost:0")

	_, err := provider.Connect(t.Context(), 3000)
	if err == nil {
		t.Errorf("expected error from Connect() with unreachable server, got nil")
	}
}

func TestSelfHosted_Connect_ServerError(t *testing.T) {
	// 1. fake server listener
	ln, err := net.Listen("tcp", "127.0.0.1:0")

	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	defer ln.Close()

	// 2. accept connection and reply with an error json
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		resp := protocol.TunnelResponse{
			Error: "subdomain already taken",
		}

		_ = json.NewEncoder(conn).Encode(resp)
	}()

	// 3. connect to the fake server
	provider := NewSelfHosted(ln.Addr().String())
	_, err = provider.Connect(t.Context(), 8080)

	// 4. verify error returned
	if err == nil || !strings.Contains(err.Error(), "subdomain already taken") {
		t.Errorf("'expected 'subdomain already taken' error, got: %v", err)
	}

}

func TestSelfHosted_ProxyRequest(t *testing.T) {
	// 1. Start a dummy local app
	localApp := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello from local app"))
	}))

	defer localApp.Close()

	localPort := localApp.Listener.Addr().(*net.TCPAddr).Port

	// 2. create an in-memory pipe simalating the tunnel
	tunnelClient, tunnelServer := net.Pipe()
	defer tunnelClient.Close()

	p := NewSelfHosted("localhost:0")

	// 3. start proxying in the bg
	go p.proxyRequest(t.Context(), tunnelServer, localPort)

	//4.  Send and http Get request to the tunnel
	req := "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n"
	if _, err := tunnelClient.Write([]byte(req)); err != nil {
		t.Fatalf("failed to write request: %v", err)
	}

	// 5. Read the response forwarded back
	buf := make([]byte, 1024)
	n, err := tunnelClient.Read(buf)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}
	if !strings.Contains(string(buf[:n]), "hello from local app") {
		t.Errorf("expected response containing 'hello from local app', got: %s", string(buf[:n]))
	}
}

func TestSelfHosted_ProxyRequest_LocalOffline(t *testing.T) {
	tunnelClient, tunnelServer := net.Pipe()
	defer tunnelClient.Close()

	p := NewSelfHosted("localhost:0")

	// Proxy to an unused port where no server is listening
	go p.proxyRequest(t.Context(), tunnelServer, 59999)

	buf := make([]byte, 1024)
	n, err := tunnelClient.Read(buf)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}
	if !strings.Contains(string(buf[:n]), "502 Bad Gateway") {
		t.Errorf("expected response containing '502 Bad Gateway', got: %s", string(buf[:n]))
	}
}

func TestSelfHosted_ConcurrentCloseAndGetters(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c2.Close()

	p := NewSelfHosted("localhost:7890")
	p.conn = c1
	p.connected = true
	p.publicURL = "http://test.example.com"

	var wg sync.WaitGroup
	wg.Add(2)

	// Goroutine 1: Concurrent reader polling status
	go func() {
		defer wg.Done()
		for range 500 {
			_ = p.IsConnected()
			_ = p.PublicURL()
		}
	}()

	// Goroutine 2: Concurrent writter shutting down
	go func() {
		defer wg.Done()
		_ = p.Close()
	}()

	wg.Wait()

}
