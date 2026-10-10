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

	httpClientTimeout = 10 * time.Second
	tcpDialTimeout    = 10 * time.Second
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
		httpClient:        httpClient,
		serverAPIEndpoint: localtunnelAPI,
		tunnelHost:        localTunnelTCPHost,
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
		// Start handling this connection
		go lt.handleConnection(lt.ctx)
	}

	return nil
}

// dialTunnel creates a single TCP connection to the localtunnel server.
func (lt *localTunnel) dialTunnel(ctx context.Context) (net.Conn, error) {
	address := net.JoinHostPort(lt.tunnelHost, strconv.Itoa(lt.tunnelPort)) //IPv6 safe

	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, fmt.Errorf("local dial failed: %w", err)
	}
	return conn, nil
}

// handleConnection processes traffic from one tunnel connection
func (lt *localTunnel) handleConnection(ctx context.Context) {

	for {
		if ctx.Err() != nil {
			return
		}
		tunnelConn, err := lt.dialTunnel(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			time.Sleep(500 * time.Millisecond) // apply backoff delay
			continue
		}

		// tunnelConn is established, start proxying
		_ = lt.proxyRequest(ctx, tunnelConn)

		// tunnel connection terminated, close it
		tunnelConn.Close()
	}

}

// proxyRequest forwards data between the tunnel connection and the local server.
func (lt *localTunnel) proxyRequest(ctx context.Context, tunnelConn net.Conn) error {
	// connect to local server
	localAddr := fmt.Sprintf("127.0.0.1:%d", lt.localPort)

	var dialer net.Dialer
	localConn, err := dialer.DialContext(ctx, "tcp", localAddr)
	if err != nil {
		return fmt.Errorf("local dial failed: %w", err)
	}
	defer localConn.Close()

	var wg sync.WaitGroup
	// Forward request: when tunnel finishes, close localConn
	wg.Go(func() {
		_, _ = io.Copy(localConn, tunnelConn)
		localConn.Close()
	})

	// Forward response: when local server finishes (EOF), close tunnelConn
	// which immediately unblocks the tunnel reader above
	wg.Go(func() {
		_, _ = io.Copy(tunnelConn, localConn)
		tunnelConn.Close()
	})

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
