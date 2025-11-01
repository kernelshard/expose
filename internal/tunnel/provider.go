package tunnel

import "context"

// Provider is an interface for tunnel service providers.
// It defines the methods required to establish and manage a tunnel.
type Provider interface {
	// Connect establishes a tunnel to the specified local port and returns the public URL.
	Connect(ctx context.Context, localPort int) (string, error)

	// Close disconnect closes the tunnel connection & cleans up resources.
	Close() error

	// IsConnected returns true if the tunnel is currently active.
	IsConnected() bool

	// PublicURL returns the public URL of the tunnel.
	PublicURL() string

	// Name of the provider (metadata)
	Name() string // "localtunnel", "ngrok", etc.
}
