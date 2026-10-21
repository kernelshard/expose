package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/kernelshard/expose/internal/pool"
	"github.com/kernelshard/expose/internal/server/protocol"
)

// Server manages the self-hosted tunnel server.
// holds the state of the server and handles incoming connections.
type Server struct {
	domain            string
	controlPort       int // port for client connections
	publicPort        int // port for public http connections
	tunnels           sync.Map
	controlTimeout    time.Duration // after this duration of no client activity, connection is closed
	handshakeTimeout  time.Duration
	maxHandshakeBytes int
	controlListener   net.Listener
	publicListener    net.Listener
	ready             chan struct{}
}

const (
	defaultDataConnBuffer   = 10
	defaultControlTimeout   = 60 * time.Second // after this duration of no client activity, connection is closed
	defaultHandshakeTimeout = 5 * time.Second
	maxHandshakeBytes       = 4 * 1024 // 4kb max handshake
)

// ClientConnection represents a connected client
type ClientConnection struct {
	conn      net.Conn
	subdomain string
	dataConns chan net.Conn
}

// NewServer creates a new server.
func NewServer(domain string, controlPort, publicPort int) *Server {
	return &Server{
		domain:            domain, // e.g. localtunnel.me
		controlPort:       controlPort,
		publicPort:        publicPort,
		controlTimeout:    defaultControlTimeout,
		handshakeTimeout:  defaultHandshakeTimeout,
		maxHandshakeBytes: maxHandshakeBytes,
		ready:             make(chan struct{}), // Used by Listeners to signal readiness.
	}
}

// startControlPlane accepts incoming client connections (registrations and data sockets)
// and dispatches each to handleControlConnection in a separate goroutine.
func (s *Server) startControlPlane() error {
	fmt.Printf("Control plane listening on %s\n", s.controlListener.Addr().String())

	for {
		conn, err := s.controlListener.Accept()
		if err != nil {
			if s.isClosed(err) {
				return nil // Graceful shutdown
			}
			return err
		}
		go s.handleControlConnection(conn)
	}
}

// startPublicServer listens for incoming public http connections from clients.
func (s *Server) startPublicServer() error {
	fmt.Printf("Public server listening on %s\n", s.publicListener.Addr().String())

	// Use Go HTTP server
	return http.Serve(s.publicListener, s)
}

// isClosed checks if the error is due to listener being closed.
func (s *Server) isClosed(err error) bool {
	return errors.Is(err, net.ErrClosed)
}

// registerSubdomain reserves a custom or random subdomain for a new client.
// It initializes the ClientConnection with a dedicated dataConns buffer
// and registers it in the tunnels map.
func (s *Server) registerSubdomain(conn net.Conn, requested string) (string, error) {
	subdomain := strings.ToLower(strings.TrimSpace(requested))
	if subdomain != "" {
		if err := ValidateSubdomain(subdomain); err != nil {
			return "", err
		}

		client := ClientConnection{
			conn:      conn,
			subdomain: subdomain,
			dataConns: make(chan net.Conn, defaultDataConnBuffer),
		}
		if _, loaded := s.tunnels.LoadOrStore(subdomain, client); loaded {
			return "", fmt.Errorf("subdomain '%s' is already in use", subdomain)
		}
		return subdomain, nil
	}

	// generate random subdomain, retry 5 times in case of collision
	for range 5 {
		sub, err := GenerateRandomSubdomain(defaultSubdomainLength)
		if err != nil {
			return "", err
		}
		client := ClientConnection{
			conn:      conn,
			subdomain: sub,
			dataConns: make(chan net.Conn, defaultDataConnBuffer),
		}
		if _, loaded := s.tunnels.LoadOrStore(sub, client); !loaded {
			return sub, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique subdomain")
}

// handleControlConnection processes an incoming client socket.
// It handles two distinct connection types:
//  1. Data connection (TypeData): stores the socket in the tenant's dataConns buffer
//     for proxying incoming public HTTP visitors.
//  2. Control connection (default): registers the subdomain, replies with the public URL,
//     and stays open to monitor client lifecycle and detect disconnects.
func (s *Server) handleControlConnection(conn net.Conn) {
	// 1. Read the handshake request
	_ = conn.SetReadDeadline(time.Now().Add(s.handshakeTimeout))
	limitedReader := io.LimitReader(conn, int64(s.maxHandshakeBytes))
	var req protocol.TunnelRequest
	if err := json.NewDecoder(limitedReader).Decode(&req); err != nil {
		conn.Close()
		return
	}

	// reset the read deadline
	_ = conn.SetReadDeadline(time.Time{})

	// Phase 1: Data socket - park in tenant's buffer and exit
	if req.Type == protocol.TypeData {
		val, ok := s.tunnels.Load(req.Subdomain)
		if !ok {
			conn.Close() // Subdomain not registered, reject
			return
		}

		client := val.(ClientConnection)
		select {
		case client.dataConns <- conn:
			return // Parked in tenant's pool
		default:
			conn.Close() // Buffer full (10 sockets already waiting), drop excess
			return
		}
	}

	// Phase 2: Control socket - register new tunnel
	subdomain, err := s.registerSubdomain(conn, req.Subdomain)
	if err != nil {
		_ = json.NewEncoder(conn).Encode(protocol.TunnelResponse{Error: err.Error()})
		conn.Close()
		return
	}

	defer s.tunnels.Delete(subdomain) // cleanup when connection closes/done

	// 4. Send Response
	publicURL := fmt.Sprintf("http://%s.%s:%d", subdomain, s.domain, s.publicPort)
	resp := protocol.TunnelResponse{
		Subdomain: subdomain,
		PublicURL: publicURL,
	}
	if err := json.NewEncoder(conn).Encode(resp); err != nil {
		conn.Close()
		return
	}

	buf := make([]byte, 1024)

	// keeping connection alive to detect disconnects
	// for other cases like ping we try to read again to check
	// that the connection is still alive
	for {
		_ = conn.SetReadDeadline(time.Now().Add(s.controlTimeout))
		_, err = conn.Read(buf)
		if err != nil {
			break
		}
	}

	conn.Close()

}

// Start runs the server. It blocks until ctx is cancelled.
func (s *Server) Start(ctx context.Context) error {

	// 1. Bind the control listener
	ctrlAddr := fmt.Sprintf(":%d", s.controlPort)
	ctrlLn, err := net.Listen("tcp", ctrlAddr)
	if err != nil {
		return fmt.Errorf("failed to bind control server to port %d: %w", s.controlPort, err)
	}
	s.controlListener = ctrlLn

	// 2. Bind public listener
	pubAddr := fmt.Sprintf(":%d", s.publicPort)
	pubLn, err := net.Listen("tcp", pubAddr)

	if err != nil {
		s.controlListener.Close()
		return fmt.Errorf("failed to bind public server to port %d: %w", s.publicPort, err)
	}
	s.publicListener = pubLn

	// 3. Signal ready
	close(s.ready)

	// 4. Setup error channel to catch startup/runtime errors
	errChan := make(chan error, 2)

	// start control plane
	go func() {

		if err := s.startControlPlane(); err != nil {
			errChan <- err
		}
	}()

	// start public server
	go func() {
		if err := s.startPublicServer(); err != nil {
			errChan <- err
		}
	}()

	// Wait for either error or ctx to be done
	select {
	case err := <-errChan:
		_ = s.Stop()
		return err
	case <-ctx.Done():
		return s.Stop()
	}
}

// ServeHTTP enables the http package to serve the tunnels. It is called by the http.Serve
// and is responsible for forwarding the request to the client connection.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Grab the first available data connection
	subdomain := extractSubdomain(r.Host, s.domain)
	if err := ValidateSubdomain(subdomain); err != nil {
		http.Error(w, "Invalid subdomain", http.StatusBadRequest)
		return
	}

	val, ok := s.tunnels.Load(subdomain)
	if !ok {
		http.Error(w, "502 Bad Gateway", http.StatusBadGateway)
		return
	}
	client := val.(ClientConnection)
	var dataConn net.Conn

	select {
	case dataConn = <-client.dataConns:
		// got one
	case <-time.After(5 * time.Second):
		http.Error(w, "504 Gateway Timeout", http.StatusGatewayTimeout)
		return
	case <-r.Context().Done():
		return
	}

	defer dataConn.Close()

	// Hijack the underlying TCP connection from net/http.
	//
	// Normally, net/http owns the connection lifecycle and expects us
	// to write an HTTP response using ResponseWriter.
	//
	// By calling Hijack(), we are telling net/http:
	// "Stop managing this connection. Give me the raw TCP socket.
	//  I will handle all reads/writes manually from now on."
	//
	// After this call:
	// - We are no longer in HTTP abstraction land.
	// - We must handle the raw TCP stream ourselves.
	// - net/http will NOT send headers or manage keep-alive.
	//
	// Hijack only works for HTTP/1.x.
	// It does NOT work for HTTP/2.
	hj, ok := w.(http.Hijacker)
	if !ok {
		// we will send internal error to client with code and hide the details
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Hijack returns:
	// 1) The raw TCP connection
	// 2) The buffered reader/writer (we don't use it)
	// 3) The error
	//
	// We immediately forward the request to the client connection
	clientConn, _, err := hj.Hijack()
	if err != nil {
		return
	}

	defer clientConn.Close()

	// Write the incoming HTTP request to the tunnel client
	//
	// r.Write() serializes the request back in to HTTP format:
	// METHOD /path HTTP/1.1
	// Headers: ...
	// Body: ...
	//
	// r.Write(dataConn) sends the request to the tunnel client
	// - This re-encodes the parsed request into bytes.
	// - If this fails, the tunnel client will not receive anything.
	// - At this point we are manually forwarding HTTP as raw bytes.
	if err := r.Write(dataConn); err != nil {
		return
	}

	// Forward the response from the tunnel client to the client
	// - This copies the raw bytes from the tunnel client to the client
	// - If this fails, the client will not receive anything
	go pool.Copy(clientConn, dataConn)
	pool.Copy(dataConn, clientConn)
}

// extractSubdomain parses the tenant subdomain from the incoming HTTP Host header.
// It returns an empty string if the host is invalid, an IP, or not a subdomain of baseDomain.
func extractSubdomain(host, baseDomain string) string {
	host = strings.ToLower(host)
	baseDomain = strings.ToLower(baseDomain)
	if baseDomain == "" {
		return ""
	}

	// when a request doesn't have a port, we strip it if present.
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}

	// A raw IP is not a domain name and cannot have subdomains
	if net.ParseIP(host) != nil {
		return ""
	}

	suffix := "." + baseDomain // e.g. ".example.com"
	if strings.HasSuffix(host, suffix) {
		return strings.TrimSuffix(host, suffix)
	}

	return ""
}

// Ready returns a channel that is closed when the server is ready.
func (s *Server) Ready() <-chan struct{} {
	return s.ready
}

// Stop gracefully shuts down the server.
func (s *Server) Stop() error {
	if s.controlListener != nil {
		s.controlListener.Close()
	}
	if s.publicListener != nil {
		s.publicListener.Close()
	}
	return nil
}

// ControlAddr returns the address of the control plane.
func (s *Server) ControlAddr() string {
	if s.controlListener == nil {
		return ""
	}
	return s.controlListener.Addr().String()
}

// SetControlTimeout sets the read deadline for the control connection.
func (s *Server) SetControlTimeout(d time.Duration) {
	s.controlTimeout = d
}

// SetHandshakeTimeout sets the handshake timeout.
func (s *Server) SetHandshakeTimeout(d time.Duration) {
	s.handshakeTimeout = d
}
