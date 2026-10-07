package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/kernelshard/expose/internal/server/protocol"
)

// SelfHosted implements the Provider interface that connects to self-hosted expose server.
type SelfHosted struct {
	serverAddr string // e.g: tunnel.mysite.com:7890
	subdomain  string
	conn       net.Conn
	publicURL  string
	connected  bool
	mu         sync.RWMutex
	// TODO: support concurrent requests like LocalTunnel (pre-open multiple data connections)
}

// NewSelfHosted creates a new self-hosted provider.
// publicURL and connected will be set after the first connection.
// serverAddr is the address of the control plane. eg. localhost:7890
func NewSelfHosted(serverAddr string) *SelfHosted {
	return &SelfHosted{
		serverAddr: serverAddr,
	}
}

// Name returns the name of the provider.
func (s *SelfHosted) Name() string {
	return "selfhosted"
}

// Connect establishes a connection to the self-hosted server.
// It sends a registration request and returns the public URL & set connection to
// conn to user for further use.
func (s *SelfHosted) Connect(ctx context.Context, localPort int) (string, error) {
	// 1. Connect to server control plane
	conn, err := net.Dial("tcp", s.serverAddr)
	if err != nil {
		return "", fmt.Errorf("failed to connect to server %w", err)
	}

	// 2. Send registration request
	req := protocol.TunnelRequest{Type: protocol.TypeControl} // empty = ask for random subdomain
	if err := json.NewEncoder(conn).Encode(req); err != nil {
		conn.Close()
		return "", err
	}
	// 3. Read server response (we expect a RegisterResponse) which contains the public URL
	var resp protocol.TunnelResponse
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		conn.Close()
		return "", err
	}

	if resp.Error != "" {
		conn.Close()
		return "", fmt.Errorf("server error: %s", resp.Error)
	}

	// 4. Store state
	s.mu.Lock()
	s.publicURL = resp.PublicURL
	s.subdomain = resp.Subdomain
	s.connected = true
	s.conn = conn
	s.mu.Unlock()

	// 5. Open data connections
	go s.openDataConnections(ctx, localPort)

	return s.publicURL, nil

}

// IsConnected returns true if the provider is connected to the server.
func (s *SelfHosted) IsConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.connected
}

// PublicURL returns the public URL of the tunnel. e.g format: http://subdomain.domain.com:port
func (s *SelfHosted) PublicURL() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.publicURL
}

// Close disconnects the tunnel and cleans up resources.
func (s *SelfHosted) Close() error {
	s.mu.Lock()
	s.connected = false
	s.publicURL = ""
	conn := s.conn
	s.conn = nil
	s.mu.Unlock()
	if conn != nil {
		return conn.Close()
	}
	return nil
}

// openDataConnections opens a pool of TCP connections to the localtunnel server.
// Each connection will handle incoming requests.
// TODO: for now we only open one connection, but we should open multiple connections to handle concurrent requests
func (s *SelfHosted) openDataConnections(ctx context.Context, localPort int) {
	var dialer net.Dialer
	const initialBackoff = 50 * time.Millisecond
	backoff := initialBackoff
	s.mu.RLock()
	subdomain := s.subdomain
	s.mu.RUnlock()

	// keep trying to open data connections
	for {
		conn, err := dialer.DialContext(ctx, "tcp", s.serverAddr)
		if err != nil {
			t := time.NewTimer(backoff)
			select {
			case <-ctx.Done():
				t.Stop()
				return
			case <-t.C:
				// timeout, try again with exponential backoff
			}
			backoff = min(backoff*2, time.Second) // 1 second max
			continue
		}
		backoff = initialBackoff

		// send data request
		req := protocol.TunnelRequest{Type: protocol.TypeData, Subdomain: subdomain}
		if err := json.NewEncoder(conn).Encode(req); err != nil {
			conn.Close()
			continue
		}

		go s.proxyRequest(ctx, conn, localPort)
	}
}

// proxyRequest proxies traffic bidirectionally between the tunnel and the local server.
func (s *SelfHosted) proxyRequest(ctx context.Context, tunnelConn net.Conn, localPort int) {
	defer tunnelConn.Close()

	localAddr := fmt.Sprintf("localhost:%d", localPort)
	var dialer net.Dialer
	localConn, err := dialer.DialContext(ctx, "tcp", localAddr)
	if err != nil {
		if ctx.Err() == nil {
			fmt.Printf("[selfhosted] failed to connect to localhost:%d: %v\n", localPort, err)
			resp := fmt.Sprintf(
				"HTTP/1.1 502 Bad Gateway\r\n"+
					"Content-Type: text/html\r\n"+
					"Connection: close\r\n\r\n"+
					"<html><body><h2>502 Bad Gateway</h2><p>Failed to connect to localhost:%d - is your server running?</p></body></html>\n",
				localPort,
			)
			_, _ = tunnelConn.Write([]byte(resp))
		}
		return
	}
	defer localConn.Close()

	var wg sync.WaitGroup

	// Forward request: stays open until tunnel closes or response completes
	wg.Go(func() {
		_, _ = io.Copy(localConn, tunnelConn)
		localConn.Close()
	})

	// Forward response: when local server finishes (EOF), closing tunnelConn
	// intentionally unblocks the reader above
	wg.Go(func() {
		_, _ = io.Copy(tunnelConn, localConn)
		tunnelConn.Close()
	})

	wg.Wait()
}
