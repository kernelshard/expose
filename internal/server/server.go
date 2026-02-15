package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/kernelshard/expose/internal/server/protocol"
)

// Server manages the self-hosted tunnel server.
// holds the state of the server and handles incoming connections.
type Server struct {
	domain          string
	controlPort     int // port for client connections
	publicPort      int // port for public http connections
	tunnels         sync.Map
	controlListener net.Listener
	publicListener  net.Listener
}

type ClientConnection struct {
	conn      net.Conn
	subdomain string
}

func NewServer(domain string, controlPort, publicPort int) *Server {
	return &Server{
		domain:      domain, // e.g. localtunnel.me
		controlPort: controlPort,
		publicPort:  publicPort,
	}
}

// Start runs the server. It blocks until ctx is cancelled.
func (s *Server) Start(ctx context.Context) error {
	// 1. Setyp error channel to catch startup errors
	errChan := make(chan error, 1)

	// start control plane
	go func() {
		if err := s.startControlPlane(); err != nil {
			errChan <- err
			return
		}
	}()

	// start public server
	go func() {
		if err := s.startPublicServer(); err != nil {
			errChan <- err
			return
		}
	}()
	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return s.Stop()
	}
}

// startControlPlane listens for incoming tunnel connections from clients.
func (s *Server) startControlPlane() error {
	addr := fmt.Sprintf("%d", s.controlPort)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.controlListener = ln

	fmt.Printf("Control plane listening on %s\n", addr)

	// accept incoming connections on the listener
	for {
		conn, err := ln.Accept()
		if err != nil {
			// check if error is due to shutdown(listener close)
			if s.isClosed(err) {
				return nil
			}
			return err
		}
		go s.handleControlConnection(conn)
	}
}

// startPublicServer listens for incoming public http connections from clients.
func (s *Server) startPublicServer() error {
	addr := fmt.Sprintf(":%d", s.publicPort)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	s.publicListener = ln
	fmt.Println("Public server listening on :", addr)

	// Use Go HTTP server
	return http.Serve(ln, s)
}

func (s *Server) isClosed(err error) bool {
	return err != nil && err.Error() != ""
}

// TODO: implement
func (s *Server) handleControlConnection(conn net.Conn) {
	// 1. Read Request
	var req protocol.RegisterRequest
	if err := json.NewDecoder(conn).Decode(&req); err != nil {
		// Failed to read JSON
		conn.Close()
		return
	}
	// 2. Assign Subdomain
	subdomain := req.Subdomain
	if subdomain == "" {
		subdomain, _ = GenerateRandomSubdomain()
	}

	// TODO: verify if subdomain is taken! for now let's assume it's not

	// 3. Register the subdomain with the connection
	client := ClientConnection{
		conn:      conn,
		subdomain: subdomain,
	}

	s.tunnels.Store(subdomain, client)
	defer s.tunnels.Delete(subdomain) // cleanup when connection closes/done

	// 4. Send Response
	publicURL := fmt.Sprintf("http://%s.%s:%d", subdomain, s.domain, s.publicPort)
	resp := protocol.RegisterResponse{
		Subdomain: subdomain,
		PublicURL: publicURL,
	}
	if err := json.NewEncoder(conn).Encode(resp); err != nil {
		conn.Close()
		return
	}

	/// 5. Wait for disconnect
	// TODO: We loop here to keep the connection alive.
	// If the client sends data (like a PING), we ignore it.
	// If the client disconnects, Read returns an error and we
	buf := make([]byte, 1)
	_, _ = conn.Read(buf)
	conn.Close()

}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, err := fmt.Fprintf(w, "Tunnel proxy coming soon!")
	if err != nil {
		fmt.Println("Error writing to response writer", err)
	}

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
