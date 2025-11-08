package cli

import (
	"github.com/spf13/cobra"

	"github.com/kernelshard/expose/internal/version"
)

var rootCmd = &cobra.Command{
	Use:     "expose",
	Short:   "Expose localhost to the internet",
	Long:    "Minimal CLI to expose your local dev server",
	Version: version.GetFullVersion(),
}

func Execute() error {

	// Add commands
	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newTunnelCmd())
	rootCmd.AddCommand(newConfigCmd())

	return rootCmd.Execute()
}
