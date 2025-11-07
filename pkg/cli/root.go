package cli

import (
	"fmt"
	"os"

	"github.com/Sanmoo/my-tasks/internal/app"
	"github.com/spf13/cobra"
)

// App instance shared across all commands
var App *app.App

// NewRootCmd creates the root command
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "tasks",
		Short: "A simple and powerful CLI task manager",
		Long: `Tasks is a CLI application for managing your tasks efficiently.
It supports creating, listing, completing, and organizing tasks with priorities and tags.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Initialize app for all commands
			var err error
			App, err = app.New()
			if err != nil {
				return fmt.Errorf("failed to initialize application: %w", err)
			}
			return nil
		},
	}

	// Add subcommands
	rootCmd.AddCommand(
		newAddCmd(),
		newListCmd(),
		newCompleteCmd(),
		newReopenCmd(),
		newDeleteCmd(),
		newUpdateCmd(),
		newShowCmd(),
	)

	return rootCmd
}

// Execute runs the root command
func Execute() {
	rootCmd := NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
