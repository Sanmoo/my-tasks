package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/Sanmoo/my-tasks/pkg/views"
	"github.com/spf13/cobra"
)

func newRemindCmd() *cobra.Command {
	var expiringIn string

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
				for _, rem := range p.GetActiveReminders() {
					// Filter by expiring-in if specified
					if expiringIn != "" {
						duration, err := parseDuration(expiringIn)
						if err != nil {
							return fmt.Errorf("invalid duration format: %w", err)
						}

						// Check if reminder expires within the specified duration
						if !rem.ExpiresIn(duration) {
							continue
						}
					} else if !rem.IsOverdue() {
						continue
					}

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

	cmd.Flags().StringVar(&expiringIn, "expiring-in", "", "Filter reminders that will expire within the specified duration (e.g., 2h, 30m, 1d)")

	return cmd
}

// parseDuration parses a duration string with support for days (d) in addition to standard time units
func parseDuration(s string) (time.Duration, error) {
	// Handle days separately since time.ParseDuration doesn't support "d"
	if len(s) > 1 && s[len(s)-1] == 'd' {
		days, err := time.ParseDuration(s[:len(s)-1] + "h")
		if err != nil {
			return 0, err
		}
		return days * 24, nil
	}

	return time.ParseDuration(s)
}
