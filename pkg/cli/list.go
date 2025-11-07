package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/Sanmoo/my-tasks/internal/task"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var (
		status   string
		priority int
		tags     []string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all tasks",
		Long:  "List all tasks with optional filtering by status, priority, or tags",
		RunE: func(cmd *cobra.Command, args []string) error {
			filter := task.Filter{}

			if status != "" {
				s := task.Status(status)
				filter.Status = &s
			}

			if priority > 0 {
				filter.Priority = &priority
			}

			if len(tags) > 0 {
				filter.Tags = tags
			}

			tasks, err := App.TaskService.ListTasks(context.Background(), filter)
			if err != nil {
				return fmt.Errorf("failed to list tasks: %w", err)
			}

			if len(tasks) == 0 {
				pterm.Info.Println("No tasks found")
				return nil
			}

			// Create table data
			tableData := pterm.TableData{
				{"ID", "Project", "Phase", "Priority", "Title", "Tags"},
			}

			for _, t := range tasks {
				priorityStr := fmt.Sprintf("%d", t.Priority)
				switch {
				case t.Priority >= 4:
					priorityStr = pterm.Red(priorityStr)
				case t.Priority == 3:
					priorityStr = pterm.Yellow(priorityStr)
				default:
					priorityStr = pterm.Gray(priorityStr)
				}

				tagsStr := "-"
				if len(t.Tags) > 0 {
					tagsStr = strings.Join(t.Tags, " ")
				}

				// Truncate ID for display
				shortID := t.ID
				if len(shortID) > 8 {
					shortID = shortID[:8]
				}

				tableData = append(tableData, []string{
					shortID,
					t.Project,
					t.Phase,
					priorityStr,
					t.Title,
					tagsStr,
				})
			}

			// Render table
			pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()

			return nil
		},
	}

	cmd.Flags().StringVarP(&status, "status", "s", "", "Filter by status (pending/completed)")
	cmd.Flags().IntVarP(&priority, "priority", "p", 0, "Filter by priority (1-5)")
	cmd.Flags().StringSliceVarP(&tags, "tags", "t", []string{}, "Filter by tags (comma-separated)")

	return cmd
}
