package protocol

// RegisterRequest is sent by the client to register a tunnel.
type RegisterRequest struct {
	// Subdomain is the desired subdomain. If empty, server will assign a random one.
	Subdomain string `json:"subdomain,omitempty"`
}

// RegisterResponse is sent by the server after registration.
type RegisterResponse struct {
	// Subdomain is the assigned subdomain.
	Subdomain string `json:"subdomain"`
	// PublicURL is the full URL to access the tunnel.
	PublicURL string `json:"public_url"`
	// Error is set if registration failed.
	Error string `json:"error,omitempty"`
}
