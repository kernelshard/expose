package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/kernelshard/expose/internal/tunnel"
)

const (
	localTunnelProviderName = "LocalTunnel"
	localtunnelAPI          = "https://localtunnel.me"
	localTunnelTCPHost      = "localtunnel.me"
	// maximum concurrent connections allowed for us,
	// override if tunnel api sends their limit
	clientMaxConn = 10

	httpClientTimeout    = 10 * time.Second
	tcpDialTimeout       = 10 * time.Second
	localDialTimeOut     = 4 * time.Second
	proxyDeadlineTimeOut = 30 * time.Second
)

// localTunnel implements the Provider interface for localtunnel.me
// It manages the lifecycle of a tunnel connection.
// It maintains a pool of TCP connections to handle incoming requests.
// It forwards traffic from the tunnel to the local server running on localPort & vice versa.
type localTunnel struct {
	publicURL      string
	localPort      int
	tunnelPort     int
	tunnelHost     string
	connected      bool
	mu             sync.RWMutex
	connections    []net.Conn // connection pool
	maxConnections int
	ctx            context.Context
	cancel         context.CancelFunc

	// HTTP client for API calls, reusable
	httpClient *http.Client
	// api endpoint string, it's configurable for testing
	serverAPIEndpoint string
}

// TunnelInfo is the response model from localtunnel server when establishing a tunnel.
type TunnelInfo struct {
	ID      string `json:"id"`
	URL     string `json:"url"`
	Port    int    `json:"port"`
	MaxConn int    `json:"max_conn_count"`
}

// NewLocalTunnel creates a new localTunnel provider instance.
func NewLocalTunnel(httpClient *http.Client) tunnel.Provider {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: httpClientTimeout}
	}

	return &localTunnel{
		connections:       make([]net.Conn, 0, clientMaxConn),
		httpClient:        httpClient,
		serverAPIEndpoint: localtunnelAPI,
	}
}

// Connect establishes tunnel to localtunnel.me
func (lt *localTunnel) Connect(ctx context.Context, localPort int) (string, error) {
	lt.mu.Lock()
	lt.localPort = localPort
	lt.ctx, lt.cancel = context.WithCancel(ctx)
	lt.mu.Unlock()

	// Step 1: Request tunnel from the localtunnel.me
	info, err := lt.requestTunnel(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to request tunnel: %w", err)
	}

	lt.mu.Lock()
	lt.publicURL = info.URL
	lt.tunnelPort = info.Port
	lt.tunnelHost = localTunnelTCPHost

	// set maxConnections allowed to open
	if info.MaxConn > 0 {
		// Take minimum: respect both server limit and our limit
		lt.maxConnections = min(info.MaxConn, clientMaxConn)
	} else {
		// Server didn't specify, use our default
		lt.maxConnections = clientMaxConn
	}

	lt.mu.Unlock()

	// Step 2: Open TCP connection pool which
	// - connects to localtunnel server
	// - handles incoming requests and forwards to local server
	// - forwards responses back to tunnel
	// We open multiple connections to handle concurrent requests
	if err := lt.openConnections(); err != nil {
		return "", fmt.Errorf("failed to open connections: %w", err)
	}

	lt.mu.Lock()
	lt.connected = true
	lt.mu.Unlock()

	return info.URL, nil

}

// requestTunnel request a tunnel from localtunnel.me API and returns the TunnelInfo.
// we make an HTTP GET request to localtunnel.me/?new
// localtunnel.me opens a tcp port for us and responds with the port
// and url info(to be used for accessing the local server)
func (lt *localTunnel) requestTunnel(ctx context.Context) (*TunnelInfo, error) {
	localTunnelReqURL := lt.serverAPIEndpoint + "/?new"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, localTunnelReqURL, nil)

	if err != nil {
		return nil, err
	}

	// Perform the HTTP request to localtunnel.me
	resp, err := lt.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("status %d:%s", resp.StatusCode, string(body))
	}

	// decode response body to TunnelInfo
	var info TunnelInfo
	err = json.NewDecoder(resp.Body).Decode(&info)
	if err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}
	return &info, nil
}

// openConnections opens a pool of TCP connections to the localtunnel server.
func (lt *localTunnel) openConnections() error {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	for i := 0; i < lt.maxConnections; i++ {
		// create tunnel connection to the upstream server & store in pool
		// each connection will handle incoming requests
		conn, err := lt.dialTunnel()
		if err != nil {
			// Close any connections we already opened
			// TODO: can do retry here instead of failing immediately
			lt.closeAllConnections()
			return fmt.Errorf("connection %d failed: %w", i, err)
		}
		// it used to close connections later
		lt.connections = append(lt.connections, conn)

		// Start handling this connection
		go lt.handleConnection(conn)
	}

	return nil
}

// dialTunnel creates a single TCP connection to the localtunnel server.
func (lt *localTunnel) dialTunnel() (net.Conn, error) {
	address := net.JoinHostPort(lt.tunnelHost, strconv.Itoa(lt.tunnelPort)) //IPv6 safe
	conn, err := net.DialTimeout("tcp", address, localDialTimeOut)

	if err != nil {
		return nil, err
	}
	return conn, nil
}

// closeAllConnections closes all existing TCP connections
func (lt *localTunnel) closeAllConnections() {
	for _, conn := range lt.connections {
		if conn != nil {
			_ = conn.Close()
		}
	}

	lt.connections = lt.connections[:0]
}

// handleConnection processes traffic from one tunnel connection
func (lt *localTunnel) handleConnection(tunnelConn net.Conn) {
	defer tunnelConn.Close()

	for {
		select {
		// run until context is done means user does Ctrl+C or Close() is called
		case <-lt.ctx.Done():
			return
		default:
			// Read request from tunnel
			// Forward to localhost
			// Write response back
			// TODO: Use connection pool instead of dialing on every request
			if err := lt.proxyRequest(tunnelConn); err != nil {
				if lt.ctx.Err() != nil {
					return // Shutting down
				}
				// Connection closed or error, exit this handler
				fmt.Printf("[localtunnel] connection error: %v\n", err)
				return
			}
		}
	}
}

// proxyRequest forwards data between the tunnel connection and the local server.
func (lt *localTunnel) proxyRequest(tunnelConn net.Conn) error {
	// connect to local server
	localAddr := fmt.Sprintf("localhost:%d", lt.localPort)
	localConn, err := net.DialTimeout("tcp", localAddr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("local dial failed: %w", err)
	}
	defer localConn.Close()

	// Set deadlines, it helps to avoid hanging connections
	// e.g: if either side doesn't respond in time, the copy will end
	_ = tunnelConn.SetDeadline(time.Now().Add(proxyDeadlineTimeOut))
	_ = localConn.SetDeadline(time.Now().Add(proxyDeadlineTimeOut))

	// Start bidirectional copy
	// mental model: copy(blocking ops) the data from tunnel to local and
	//local to tunnel concurrently when either side closes, the copy ends
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(localConn, tunnelConn)
	}()

	go func() {
		defer wg.Done()
		io.Copy(tunnelConn, localConn)
	}()

	wg.Wait()
	return nil

}

// Close terminates the tunnel
func (lt *localTunnel) Close() error {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	if lt.cancel != nil {
		lt.cancel()

	}

	lt.closeAllConnections()
	lt.connected = false
	return nil
}

// IsConnected returns true if tunnel is active
func (lt *localTunnel) IsConnected() bool {
	lt.mu.RLock()
	defer lt.mu.RUnlock()
	return lt.connected
}

func (lt *localTunnel) PublicURL() string {
	lt.mu.RLock()
	defer lt.mu.RUnlock()
	return lt.publicURL
}

func (lt *localTunnel) Name() string {
	return localTunnelProviderName
}
