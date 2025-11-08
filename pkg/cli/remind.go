package cli

import (
	"context"
	"fmt"

	"github.com/Sanmoo/my-tasks/pkg/views"
	"github.com/spf13/cobra"
)

func newRemindCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remind",
		Short: "List all tasks from all scannable projects with due reminders",
		Long:  "List all tasks from all scannable projects with due reminders",
		Args:  cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			projects, err := App.TaskService.GetAllProjects(context.Background())
			if err != nil {
				return err
			}

			viewReminders := []views.Reminder{}
			for _, p := range projects {
				for _, rem := range p.GetOverdueReminders() {
					viewReminders = append(viewReminders, views.Reminder{
						Project: p.GetName(),
						Label:   fmt.Sprintf("Task: %s. Reminder: %s.", rem.TaskTitle, rem.Label),
					})
				}
			}

			reminderPanel := views.ReminderPanel{Reminders: viewReminders}

			reminderPanel.Render()

			return nil
		},
	}

	return cmd
}
