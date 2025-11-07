package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/Sanmoo/my-tasks/internal/task"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func newShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show [task-id]",
		Short: "Show detailed information about a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := args[0]

			t, err := App.TaskService.GetTask(context.Background(), taskID)
			if err != nil {
				return fmt.Errorf("failed to get task: %w", err)
			}

			// Display task details
			pterm.DefaultSection.Println("Task Details")

			details := [][]string{
				{"ID", t.ID},
				{"Title", t.Title},
				{"Project", t.Project},
				{"Phase", t.Phase},
				{"Status", string(t.Status)},
			}

			if len(t.Tags) > 0 {
				details = append(details, []string{"Tags", strings.Join(t.Tags, " ")})
			}

			if t.DueDate != nil {
				details = append(details, []string{"Due Date", t.DueDate.Format("2006-01-02 15:04:05")})
			}

			if t.Status == task.StatusCompleted {
				details = append(details, []string{"Completed"})
			}

			for _, detail := range details {
				pterm.Printf("%s: %s\n", pterm.Bold.Sprint(detail[0]), detail[1])
			}

			// Display comments
			if len(t.Comments) > 0 {
				pterm.Println()
				pterm.DefaultSection.Println("Comments")
				for _, comment := range t.Comments {
					pterm.Printf("  - %s\n", comment)
				}
			}

			// Display reminders
			if len(t.Reminders) > 0 {
				pterm.Println()
				pterm.DefaultSection.Println("Reminders")
				for _, reminder := range t.Reminders {
					status := "Pending"
					if reminder.Acknowledged {
						status = "Acknowledged"
					}
					pterm.Printf("  - %s (%s)\n", reminder.Time.Format("2006-01-02 15:04:05"), status)
				}
			}

			return nil
		},
	}

	return cmd
}
