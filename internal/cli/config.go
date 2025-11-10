package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kernelshard/expose/internal/config"
)

// newConfigCmd creates the 'config' command
func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage expose configuration",
		Long:  "View and manage expose configuration values",
	}

	//
	cmd.AddCommand(newConfigListCmd())
	cmd.AddCommand(newConfigGetCmd())

	return cmd
}

// newConfigListCmd creates the 'config list' command
func newConfigListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all configuration values",
		RunE:  runConfigList,
	}
	return cmd
}

// newConfigGetCmd creates the 'config get' command
// e.g. expose config get <key>
func newConfigGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get a specific configuration value",
		Args:  cobra.ExactArgs(1),
		RunE:  runConfigGet,
	}
}

// runConfigList handles the 'config list' command
func runConfigList(_ *cobra.Command, args []string) error {
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("config not found (run 'expose init' first): %w", err)
	}
	values := cfg.List()
	for key, value := range values {
		fmt.Printf("%s: %v\n", key, value)
	}
	return nil
}

// runConfigGet handles the 'config get <key>' command
func runConfigGet(_ *cobra.Command, args []string) error {
	key := args[0]
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("config not found (run 'expose init' first): %w", err)
	}
	val, err := cfg.Get(key)
	if err != nil {
		return err
	}
	fmt.Println(val)
	return nil
}
