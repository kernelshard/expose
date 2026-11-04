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
	Subdomain      string `json:"subdomain"`
	PublicURL      string `json:"public_url"`
	MaxConnections int    `json:"max_connections,omitempty"` // server-advertised maximum data connections for this tunnel.
	Error          string `json:"error,omitempty"`
}
