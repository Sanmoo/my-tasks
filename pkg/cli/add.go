package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newAddCmd() *cobra.Command {
	var (
		project  string
		phase    string
		priority int
		tags     []string
		comments []string
	)

	cmd := &cobra.Command{
		Use:   "add [title]",
		Short: "Add a new task",
		Long:  "Add a new task with project, phase, priority, tags, and comments",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			title := strings.Join(args, " ")

			// Set defaults if not provided
			if project == "" {
				project = "Default"
			}
			if phase == "" {
				phase = "Planejado"
			}

			task, err := App.TaskService.CreateTask(
				context.Background(),
				title,
				project,
				phase,
				priority,
				tags,
				comments,
			)
			if err != nil {
				return fmt.Errorf("failed to create task: %w", err)
			}

			fmt.Printf("Task created successfully!\n")
			fmt.Printf("ID: %s\n", task.ID)
			fmt.Printf("Title: %s\n", task.Title)
			fmt.Printf("Project: %s\n", task.Project)
			fmt.Printf("Phase: %s\n", task.Phase)
			fmt.Printf("Priority: %d\n", task.Priority)
			if len(task.Tags) > 0 {
				fmt.Printf("Tags: %s\n", strings.Join(task.Tags, ", "))
			}
			if len(task.Comments) > 0 {
				fmt.Printf("Comments:\n")
				for _, comment := range task.Comments {
					fmt.Printf("  - %s\n", comment)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&project, "project", "", "Project name (default: 'Default')")
	cmd.Flags().StringVar(&phase, "phase", "", "Phase/column name (default: 'Planejado')")
	cmd.Flags().IntVarP(&priority, "priority", "p", 3, "Task priority (1-5, where 5 is highest)")
	cmd.Flags().StringSliceVarP(&tags, "tags", "t", []string{}, "Task tags (space-separated)")
	cmd.Flags().StringSliceVarP(&comments, "comments", "c", []string{}, "Task comments")

	return cmd
}
