package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/Sanmoo/my-tasks/pkg/views"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var (
		status string
		tags   []string
	)

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

			columns := []views.Column{}

			for _, project := range projects {
				for _, p := range project.GetPhases() {
					columns = append(columns, views.Column{
						Name:  p.GetName(),
						Tasks: p.GetTaskTitles(),
					})
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

	cmd.Flags().StringVarP(&status, "status", "s", "", "Filter by status (pending/completed)")
	cmd.Flags().StringSliceVarP(&tags, "tags", "t", []string{}, "Filter by tags (comma-separated)")

	return cmd
}
