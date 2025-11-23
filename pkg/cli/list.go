package cli

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/Sanmoo/my-tasks/internal/task"
	"github.com/Sanmoo/my-tasks/pkg/views"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var statuses []string
	var tags []string

	cmd := &cobra.Command{
		Use:   "list [project_names]",
		Short: "List all tasks from named projects, comma-separated",
		Long:  "List all tasks from named projects, comma-separated",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if a project name was provided
			var projectNamesOrAliases string

			if len(args) == 0 {
				projectNamesOrAliases = App.Config.DefaultProject
				if projectNamesOrAliases == "" {
					return fmt.Errorf("a project name or alias must be provided or a default one must be configured")
				}
			} else {
				projectNamesOrAliases = args[0]
			}

			projects, err := App.TaskService.GetProjectsByNamesOrAliases(context.Background(), strings.Split(projectNamesOrAliases, ","))
			if err != nil {
				return fmt.Errorf("failed to find projects with name %s: %w", projectNamesOrAliases, err)
			}

			RenderWarningsIfAny(projects)

			for _, project := range projects {
				columns := []views.Column{}

				for _, p := range project.GetPhases() {
					if len(statuses) == 0 || slices.Contains(statuses, string(p.GetAssociatedStatus())) {
						// Filter tasks by tags if specified
						filteredTasks := []string{}
						for _, task := range p.GetTasks() {
							if len(tags) == 0 || hasAnyTag(task, tags) {
								filteredTasks = append(filteredTasks, task.GetTitle())
							}
						}

						if len(filteredTasks) > 0 {
							columns = append(columns, views.Column{
								Name:  p.GetName(),
								Tasks: filteredTasks,
							})
						}
					}
				}

				kanban := views.Kanban{
					ProjectName: project.GetName(),
					Columns:     columns,
				}

				kanban.Render()
			}

			return nil
		},
	}

	cmd.Flags().StringSliceVarP(&statuses, "status", "s", []string{}, "Filter by task statuses (comma-separated). Available statuses: pending, running, scheduled, completed")
	cmd.Flags().StringSliceVarP(&tags, "tags", "t", []string{}, "Filter by task tags (comma-separated)")

	return cmd
}

// hasAnyTag checks if a task has any of the specified tags
func hasAnyTag(task *task.Task, tags []string) bool {
	taskTags := task.GetTags()
	for _, tag := range tags {
		if slices.Contains(taskTags, tag) {
			return true
		}
	}
	return false
}
