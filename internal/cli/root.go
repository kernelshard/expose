package cli

import (
	"github.com/spf13/cobra"
)

func Execute() error {
	rootCmd := &cobra.Command{
		Use:   "expose",
		Short: "Expose localhost to the internet",
		Long:  "Minimal CLI to expose your local dev server",
	}

	// Add commands
	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newTunnelCmd())

	return rootCmd.Execute()
}
