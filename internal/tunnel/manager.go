package tunnel

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

// Tunneler represents a tunnel that can be started and stopped, and
// provides a public URL once ready.
type Tunneler interface {
	Start(ctx context.Context) error
	Close() error
	Ready() <-chan struct{}
	PublicURL() string
}

// Manager manages the lifecycle of a tunneler.
type Manager struct {
	localPort int
	publicURL string
	listener  net.Listener
	server    *http.Server
	ready     chan struct{}
	mu        sync.RWMutex
}

// Ensure Manager implements Tunneler
var _ Tunneler = (*Manager)(nil)

// NewManager creates a new Manager instance.
func NewManager(port int) *Manager {
	return &Manager{
		localPort: port,
		ready:     make(chan struct{}),
	}
}

// Start initializes the tunnel and begins listening for incoming connections.
func (m *Manager) Start(ctx context.Context) error {
	// respect context cancellation; exit early if already cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Create a Listener
	listener, err := net.Listen("tcp", ":0") // Listen on any random available port
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}

	// Set the public URL and listener (concurrency-safe)
	port := listener.Addr().(*net.TCPAddr).Port
	m.mu.Lock()
	m.listener = listener
	m.publicURL = fmt.Sprintf("http://localhost:%d", port)
	m.mu.Unlock()

	// Signal that the tunnel is ready
	// closing the channel indicates readiness as per Go idioms
	// concurrency-safe
	// we can do heavy operations here before signaling readiness if needed
	// e.g., establishing connections to remote servers
	close(m.ready)

	// Create HTTP server to handle incoming requests
	server := &http.Server{
		Handler: http.HandlerFunc(m.proxyHandler),
	}

	// Set server (concurrency-safe)
	m.mu.Lock()
	m.server = server
	m.mu.Unlock()

	// Auto-close & clean up on context cancellation
	go func() {
		<-ctx.Done()
		m.Close()
	}()

	// Serve incoming connections(blocking call)
	// ends when closed from outside (e.g., via m.Close()) or context cancellation
	if err := m.server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

// Ready returns a channel that is closed when the tunnel is ready.
// before the tunnel is ready it will block on reads.
func (m *Manager) Ready() <-chan struct{} {
	return m.ready
}

// Close shuts down the tunnel and cleans up resources.
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var err error

	// Shutdown the http server if it's running
	if m.server != nil {
		err = m.server.Close()
	} else if m.listener != nil {
		err = m.listener.Close()
	}

	return err

}

// PublicURL returns the public URL of the tunnel.
// for concurrency safety we read under a lock.
func (m *Manager) PublicURL() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.publicURL
}

// proxyHandler forwards incoming HTTP requests to the local server.
// It dials the local server, forwards the request, and writes back the response.
// If any step fails, it responds with an appropriate HTTP error.
func (m *Manager) proxyHandler(w http.ResponseWriter, r *http.Request) {

	// create connection to local server
	target := fmt.Sprintf("localhost:%d", m.localPort)
	conn, err := net.DialTimeout("tcp", target, 5*time.Second)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to connect localhost:%d - is your server running?", m.localPort), http.StatusBadGateway)
		return
	}

	defer conn.Close()

	// Send request to local server
	if err := r.Write(conn); err != nil {
		http.Error(w, "Failed to forward request", http.StatusBadGateway)
		return
	}

	// Read response from local server
	resp, err := http.ReadResponse(bufio.NewReader(conn), r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read response from local server: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Copy response status code and body
	w.WriteHeader(resp.StatusCode)

	// partial response sent anyway as headers are already written
	io.Copy(w, resp.Body) // nolint:errcheck

}
