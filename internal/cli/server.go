package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/kernelshard/expose/internal/server"
)

// newServerCmd creates the 'expose server' command.
// it starts a self-hosted expose server.
func newServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Run a self-hosted expose server",
		Long: `Start a self-hosted tunnel server on your VPS.
				
		Clients can connect using:
			expose tunnel --server=yourdomain.com --provider=selfhosted
		
		Example:
			expose server --domain=tunnel.mysite.com --control-port=7890 --public-port=8080`,
		RunE: runServerCmd,
	}

	cmd.Flags().IntP("control-port", "c", 7890, "Port for client connections")
	cmd.Flags().IntP("public-port", "p", 8080, "Port for public HTTP traffic")
	cmd.Flags().StringP("domain", "d", "localhost", "Domain for public URLs")
	return cmd
}

// runServerCmd runs the  self-hosted tunnel server.
func runServerCmd(cmd *cobra.Command, _ []string) error {
	domain, _ := cmd.Flags().GetString("domain")
	controlPort, _ := cmd.Flags().GetInt("control-port")
	publicPort, _ := cmd.Flags().GetInt("public-port")

	srv := server.NewServer(domain, controlPort, publicPort)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down server...")
		cancel()
	}()

	fmt.Printf("🚀 Expose server starting...\n")
	return srv.Start(ctx)
}
