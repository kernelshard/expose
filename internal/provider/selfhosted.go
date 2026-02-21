package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net"

	"github.com/kernelshard/expose/internal/server/protocol"
)

// SelfHosted implements the Provider interface that connects to self-hosted expose server.
type SelfHosted struct {
	serverAddr string // e.g: tunnel.mysite.com:7890
	conn       net.Conn
	publicURL  string
	connected  bool
}

// NewSelfHosted creates a new self-hosted provider.
// publicURL and connected will be set after the first connection.
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
	req := protocol.RegisterRequest{} // empty = ask for random subdomain
	if err := json.NewEncoder(conn).Encode(req); err != nil {
		conn.Close()
		return "", err
	}
	// 3. Read server response (we expect a RegisterResponse) which contains the public URL
	var resp protocol.RegisterResponse
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
	s.connected = true
	s.conn = conn

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
