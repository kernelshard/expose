package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/kernelshard/expose/internal/config"
	"github.com/kernelshard/expose/internal/tunnel"
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

	// Create manager
	mgr := tunnel.NewManager(port)

	// context with signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// handle Ctrl+C, kill pid etc
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// waiting to read from channel is blocking ops, so wait in bg.
	go func() {
		<-sigChan
		fmt.Println("\n\nShutting down...")
		cancel()
	}()

	// start in background
	errChan := make(chan error, 1)
	go func() {
		errChan <- mgr.Start(ctx)
	}()

	// wait for ready
	<-mgr.Ready()

	// Show info
	fmt.Printf("🚀 Starting tunnel for localhost:%d\n\n", port)
	fmt.Printf("✓ Public URL:   %s\n", mgr.PublicURL())
	fmt.Printf("✓ Forwarding to: http://localhost:%d\n\n", port)
	fmt.Println("Press Ctrl+C to stop")

	return <-errChan
}
