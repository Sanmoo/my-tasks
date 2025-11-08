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

			RenderWarningsIfAny(projects)

			reminderPanel := views.NewReminderPanel()
			for _, p := range projects {
				for _, rem := range p.GetOverdueReminders() {
					reminderPanel.AddReminder(views.Reminder{
						Project: p.GetName(),
						Label:   fmt.Sprintf("%s: %s.", rem.Label, rem.TaskTitle),
					})
				}
			}

			reminderPanel.Render()

			return nil
		},
	}

	return cmd
}
