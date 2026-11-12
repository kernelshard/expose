package cli

import (
	"github.com/spf13/cobra"

	"github.com/kernelshard/expose/internal/version"
)

// NewRootCmd creates and initializes the root cobra command with all subcommands.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "expose",
		Short:   "Expose localhost to the internet",
		Long:    "Minimal CLI to expose your local dev server",
		Version: version.GetFullVersion(),
	}

	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newTunnelCmd())
	cmd.AddCommand(newServerCmd())
	cmd.AddCommand(newConfigCmd())

	return cmd
}

func Execute() error {
	return NewRootCmd().Execute()
}
