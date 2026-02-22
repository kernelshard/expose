package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
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
	s.publicURL = resp.PublicURL
	s.subdomain = resp.Subdomain
	s.connected = true
	s.conn = conn

	// 5. Open data connections
	go s.openDataConnections(ctx, localPort)

	return s.publicURL, nil

}

// IsConnected returns true if the provider is connected to the server.
func (s *SelfHosted) IsConnected() bool {
	return s.connected
}

// PublicURL returns the public URL of the tunnel. e.g format: http://subdomain.domain.com:port
func (s *SelfHosted) PublicURL() string {
	return s.publicURL
}

// Close disconnects the tunnel and cleans up resources.
func (s *SelfHosted) Close() error {
	s.connected = false
	s.publicURL = ""
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}

// openDataConnections opens a pool of TCP connections to the localtunnel server.
// Each connection will handle incoming requests.
// TODO: for now we only open one connection, but we should open multiple connections to handle concurrent requests
func (s *SelfHosted) openDataConnections(ctx context.Context, localPort int) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		conn, err := net.Dial("tcp", s.serverAddr)
		if err != nil {
			continue
		}
		json.NewEncoder(conn).Encode(protocol.TunnelRequest{
			Type:      protocol.TypeData,
			Subdomain: s.subdomain,
		})

		go s.proxyRequest(conn, localPort)
	}
}

// proxyRequest proxies traffic from the tunnel connection to the local server and vice versa.
func (s *SelfHosted) proxyRequest(tunnelConn net.Conn, localPort int) {
	defer tunnelConn.Close()

	localAddr := fmt.Sprintf("localhost:%d", localPort)
	localConn, err := net.DialTimeout("tcp", localAddr, 5*time.Second)
	if err != nil {
		return
	}
	defer localConn.Close()

	// a. Start a background goroutine to pump data FROM local TO tunnel
	go io.Copy(localConn, tunnelConn)
	// b. Block the main thread pumping data FROM tunnel TO local
	io.Copy(tunnelConn, localConn)
}
