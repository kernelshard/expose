package cli

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/kernelshard/expose/internal/config"
	"github.com/spf13/cobra"
)

// tunnelCmd represents the 'tunnel' command in the CLI application.
func newTunnelCmd() *cobra.Command {
	var port int

	// Define the command structure and behavior
	cmd := &cobra.Command{
		Use:   "tunnel",
		Short: "Start a tunnel to expose local server",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load config
			cfg, err := config.Load("")
			if err != nil {
				return fmt.Errorf("run 'expose' init first")
			}

			if port == 0 {
				port = cfg.DefaultPort
			}

			return runTunnel(port)
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 0, "local port to expose")
	return cmd
}

// runTunnel sets up a reverse proxy to expose the local server
// on the specified port.
func runTunnel(port int) error {
	fmt.Printf("Exposing localhost:%d\n", port)
	fmt.Printf("Local URL: http://localhost:8080\n")
	fmt.Println("\nPress Ctrl+C to stop")

	// Create reverse proxy to forward requests
	target, _ := url.Parse(fmt.Sprintf("http://localhost:%d", port))
	proxy := httputil.NewSingleHostReverseProxy(target)

	// Start local server
	return http.ListenAndServe(":8080", proxy)
}
