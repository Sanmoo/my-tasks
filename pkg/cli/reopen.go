package cli

import (
	"context"
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func newReopenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reopen [task-id]",
		Short: "Reopen a completed task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := args[0]

			if err := App.TaskService.ReopenTask(context.Background(), taskID); err != nil {
				return fmt.Errorf("failed to reopen task: %w", err)
			}

			pterm.Success.Printf("Task %s reopened\n", taskID)
			return nil
		},
	}

	return cmd
}
