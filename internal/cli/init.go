package cli

import (
	"fmt"

	"github.com/kernelshard/expose/internal/config"
	"github.com/spf13/cobra"
)

// newInitCmd creates the 'init' command for initializing configuration.
func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize expose configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Init()
			if err != nil {
				return err
			}

			fmt.Printf("✓ Created .expose.yml\n")
			fmt.Printf("✓ Project: %s\n", cfg.Project)
			fmt.Printf("✓ Port: %d\n", cfg.DefaultPort)
			return nil

		},
	}
}
