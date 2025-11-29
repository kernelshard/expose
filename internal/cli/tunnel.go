package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/kernelshard/expose/internal/config"
	"github.com/kernelshard/expose/internal/provider"
	"github.com/kernelshard/expose/internal/tunnel"
)

// tunnelCmd represents the tunnel command
func newTunnelCmd() *cobra.Command {
	//Use:   "tunnel",
	//Short: "Start a tunnel to expose local server",
	//RunE:  runTunnelCmd,
	cmd := &cobra.Command{
		Use:   "tunnel",
		Short: "Expose local server via tunnel",
		RunE:  runTunnelCmd,
	}

	// Define flags
	// provider flag to specify provider e.g. expose tunnel --provider cloudflare
	cmd.Flags().StringP("provider", "P", "localtunnel", "Tunnel provider: localtunnel, cloudflare, etc. defaults to localtunnel")

	// port flag to specify local port e.g. expose tunnel --port 8080
	cmd.Flags().IntP("port", "p", 0, "Local port to expose (overrides config)")
	return cmd
}

// runTunnelCmd represents the 'tunnel' command in the CLI application.
func runTunnelCmd(cmd *cobra.Command, _ []string) error {

	// Load config
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("config not found (run 'expose init' first): %w", err)
	}

	// Get port from flag
	port, err := cmd.Flags().GetInt("port")
	if err != nil {
		return fmt.Errorf("invalid port flag %w", err)
	}

	// use config port if flag not set
	if port == 0 {
		port = cfg.Port
	}

	if port <= 0 || port > 65535 {
		return fmt.Errorf("invalid port %d (must be 1-65535)", port)
	}

	// use provider flag shorthand -P to select provider
	providerName, err := cmd.Flags().GetString("provider")
	if err != nil {
		return fmt.Errorf("invalid port %d (must be 1-65535)", port)
	}

	return runTunnel(port, providerName)
}

// runTunnel sets up a reverse proxy to expose the local server
// on the specified port.
func runTunnel(port int, providerName string) error {
	var svc *tunnel.Service

	switch providerName {
	case "cloudflare":
		svc = tunnel.NewService(provider.NewCloudFlare())
	default:
		svc = tunnel.NewService(provider.NewLocalTunnel(nil))

	}

	// Setup ctx & signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// handle Ctrl+C, kill pid etc.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// waiting to read from channel is blocking ops, so wait in bg.
	go func() {
		<-sigChan
		fmt.Println("\n\nShutting down...")
		cancel()
	}()

	// - Start  tunnel in background
	errChan := make(chan error, 1)
	go func() {
		errChan <- svc.Start(ctx, port)
	}()

	// wait for ready
	select {
	case <-svc.Ready():
		// Show info
		fmt.Printf("🚀 Tunnel[%s] started for localhost:%d\n", svc.ProviderName(), port)
		fmt.Printf("✓ Public URL: %s\n", svc.PublicURL())
		fmt.Printf("✓ Forwarding to: http://localhost:%d\n", port)
		fmt.Printf("✓ Provider: %s\n", svc.ProviderName())
		fmt.Println("Press Ctrl+C to stop")

	case err := <-errChan:
		if err != nil {
			return err
		}

	}

	// - Wait for shutdown
	<-ctx.Done()

	// - Cleanup
	if err := svc.Close(); err != nil {
		return fmt.Errorf("close failed %w", err)
	}

	fmt.Println("✓ Tunnel closed")
	return nil
}
