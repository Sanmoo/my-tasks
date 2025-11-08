package cli

import (
	"context"
	"fmt"

	"github.com/Sanmoo/my-tasks/pkg/views"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var (
		status   string
		priority int
		tags     []string
	)

	cmd := &cobra.Command{
		Use:   "list [project_name]",
		Short: "List all tasks from a specific project",
		Long:  "List all tasks from a specific project",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if a project name was provided
			if len(args) == 0 {
				return fmt.Errorf("a project name or alias must be provided")
			}

			project, err := App.TaskService.GetProjectByNameOrAlias(context.Background(), args[0])
			if err != nil {
				return fmt.Errorf("failed to find projects with name %s: %w", args[0], err)
			}

			columns := []views.Column{}

			for _, p := range *project.GetPhases() {
				for currentNode, i := p.GetTasks().Front, 1; currentNode != nil; currentNode, i = currentNode.Next, i+1 {
					columns = append(columns, views.Column{
						Name:  p.GetName(),
						Tasks: p.GetTaskTitles(),
					})
				}
			}

			kanban := views.Kanban{
				ProjectName: args[0],
				Columns:     columns,
			}

			kanban.Render()

			return nil
		},
	}

	cmd.Flags().StringVarP(&status, "status", "s", "", "Filter by status (pending/completed)")
	cmd.Flags().IntVarP(&priority, "priority", "p", 0, "Filter by priority (1-5)")
	cmd.Flags().StringSliceVarP(&tags, "tags", "t", []string{}, "Filter by tags (comma-separated)")

	return cmd
}
