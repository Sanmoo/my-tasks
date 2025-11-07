package cli

import (
	"context"
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func newCompleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "complete [task-id]",
		Short: "Mark a task as completed",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := args[0]

			if err := App.TaskService.CompleteTask(context.Background(), taskID); err != nil {
				return fmt.Errorf("failed to complete task: %w", err)
			}

			pterm.Success.Printf("Task %s marked as completed\n", taskID)
			return nil
		},
	}

	return cmd
}
