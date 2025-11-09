package cli

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/Sanmoo/my-tasks/pkg/views"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var statuses []string

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
						columns = append(columns, views.Column{
							Name:  p.GetName(),
							Tasks: p.GetTaskTitles(),
						})
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

	cmd.Flags().StringSliceVarP(&statuses, "status", "s", []string{}, "Filter by task statuses (comma-separated)")

	return cmd
}
