package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "expose",
	Short: "Expose localhost to the internet",
	Long:  "Minimal CLI to expose your local dev server",
}

func Execute() error {

	// Add commands
	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newTunnelCmd())

	return rootCmd.Execute()
}
