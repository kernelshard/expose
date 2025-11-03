package tunnel

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Service wraps a tunnel Provider and manages its lifecycle.
// It provides a uniform interface for all tunnel providers(localtunnel, ngrok etc.)
type Service struct {
	provider Provider
	ready    chan struct{}
	mu       sync.RWMutex
	started  bool
	closed   bool
}

// NewService creates a new Service instance with the given Provider.
func NewService(p Provider) *Service {
	return &Service{
		provider: p,
		ready:    make(chan struct{}),
	}
}

// Start initializes the tunnel provider and signals when ready.
func (s *Service) Start(ctx context.Context, localPort int) error {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return fmt.Errorf("tunnel already started")
	}

	if s.closed {
		s.mu.Unlock()
		return fmt.Errorf("service is closed")
	}
	s.started = true
	s.mu.Unlock()

	_, err := s.provider.Connect(ctx, localPort)
	if err != nil {
		return fmt.Errorf("failed to connect %s provider tunnel: %w", s.provider.Name(), err)
	}

	// signal that tunnel is ready to use
	close(s.ready)
	return nil

}

// Ready returns a channel that closes when the tunnel is ready.
// Useful for waiting in CLI: <-service.Ready()
func (s *Service) Ready() <-chan struct{} {
	return s.ready
}

// PublicURL returns the tunnel's public URL.
// Returns empty string if not connected.
func (s *Service) PublicURL() string {
	return s.provider.PublicURL()
}

// ProviderName returns the name of the tunnel provider.
func (s *Service) ProviderName() string {
	return s.provider.Name()
}

// IsConnected returns true if tunnel is active
func (s *Service) IsConnected() bool {
	return s.provider.IsConnected()
}

// Close terminates the tunnel and cleans up resources.
func (s *Service) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	s.mu.Unlock()

	return s.provider.Close()
}

// WaitReady waits for the tunnel to be ready with a timeout.
// Returns error if timeout exceeded or service closes
func (s *Service) WaitReady(timeout time.Duration) error {
	if s.provider.IsConnected() {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	select {
	case <-s.ready:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("tunnel readiness timeout: %w", ctx.Err())
	}
}
