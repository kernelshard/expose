package protocol

const (
	TypeControl = "control"
	TypeData    = "data"
)

// TunnelRequest is sent by the client when connecting to the server.
// Type determines the purpose: TypeControl for registration, TypeData for proxying.
type TunnelRequest struct {
	// Subdomain is the desired subdomain. If empty, server will assign a random one.
	Subdomain string `json:"subdomain,omitempty"`
	Type      string `json:"type,omitempty"` // "control" or "data"
}

// TunnelResponse is sent by the server after registration.
type TunnelResponse struct {
	// Subdomain is the assigned subdomain.
	Subdomain string `json:"subdomain"`
	// PublicURL is the full URL to access the tunnel.
	PublicURL string `json:"public_url"`
	// Error is set if registration failed.
	Error string `json:"error,omitempty"`
}
