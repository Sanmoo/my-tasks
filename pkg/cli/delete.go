package cli

import (
	"context"
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete [task-id]",
		Short: "Delete a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := args[0]

			if !force {
				result, _ := pterm.DefaultInteractiveConfirm.
					WithDefaultText(fmt.Sprintf("Are you sure you want to delete task %s?", taskID)).
					Show()
				if !result {
					pterm.Info.Println("Delete cancelled")
					return nil
				}
			}

			if err := App.TaskService.DeleteTask(context.Background(), taskID); err != nil {
				return fmt.Errorf("failed to delete task: %w", err)
			}

			pterm.Success.Printf("Task %s deleted\n", taskID)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")

	return cmd
}
