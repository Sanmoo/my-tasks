package cli

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/Sanmoo/my-tasks/internal/app"
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
		Use:   "list [project_name]",
		Short: "List all tasks or tasks from a specific project",
		Long:  "List all tasks with optional filtering by status, priority, or tags. If a project name is provided, lists only tasks from that project grouped by phase.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if a project name was provided
			if len(args) == 1 {
				return listProjectTasks(args[0])
			}

			// Original list all tasks behavior
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
					t.Title,
					tagsStr,
				})
			}

			// Render table
			err = pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
			if err != nil {
				return nil
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&status, "status", "s", "", "Filter by status (pending/completed)")
	cmd.Flags().IntVarP(&priority, "priority", "p", 0, "Filter by priority (1-5)")
	cmd.Flags().StringSliceVarP(&tags, "tags", "t", []string{}, "Filter by tags (comma-separated)")

	return cmd
}

// listProjectTasks lists all tasks from a specific project, grouped by phase
func listProjectTasks(projectName string) error {
	// Find the file containing this project
	projectFile, err := App.GetProjectFile(projectName)
	if err != nil {
		return fmt.Errorf("failed to find project: %w", err)
	}

	// Create a new app instance with the project file
	projectApp, err := app.NewWithProjectFile(projectFile)
	if err != nil {
		return fmt.Errorf("failed to load project: %w", err)
	}

	// Resolve project alias to full name
	fullProjectName := projectApp.Config.ResolveProjectName(projectName)

	// List all tasks from this project
	tasks, err := projectApp.TaskService.ListTasks(context.Background(), task.Filter{})
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}

	// Filter tasks by project and group by phase
	tasksByPhase := make(map[string][]*task.Task)
	for _, t := range tasks {
		if t.Project == fullProjectName {
			tasksByPhase[t.Phase] = append(tasksByPhase[t.Phase], t)
		}
	}

	if len(tasksByPhase) == 0 {
		pterm.Info.Printf("No tasks found in project '%s'\n", fullProjectName)
		return nil
	}

	// Sort phases alphabetically
	phases := make([]string, 0, len(tasksByPhase))
	for phase := range tasksByPhase {
		phases = append(phases, phase)
	}
	sort.Strings(phases)

	// Display project header
	pterm.DefaultHeader.WithFullWidth().
		WithBackgroundStyle(pterm.NewStyle(pterm.BgLightBlue)).
		WithTextStyle(pterm.NewStyle(pterm.FgBlack)).
		Printf("Project: %s", fullProjectName)
	fmt.Println()

	// Display tasks grouped by phase
	for _, phase := range phases {
		phaseTasks := tasksByPhase[phase]

		// Phase header
		pterm.DefaultSection.Printf("Phase: %s (%d tasks)", phase, len(phaseTasks))

		// Create table data for this phase
		tableData := pterm.TableData{
			{"ID", "Priority", "Title", "Status", "Tags"},
		}

		for _, t := range phaseTasks {
			tagsStr := "-"
			if len(t.Tags) > 0 {
				tagsStr = strings.Join(t.Tags, " ")
			}

			// Truncate ID for display
			shortID := t.ID
			if len(shortID) > 8 {
				shortID = shortID[:8]
			}

			statusStr := string(t.Status)
			if t.Status == task.StatusCompleted {
				statusStr = pterm.Green(statusStr)
			}

			tableData = append(tableData, []string{
				shortID,
				t.Title,
				statusStr,
				tagsStr,
			})
		}

		// Render table
		err = pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
		if err != nil {
			return err
		}

		fmt.Println()
	}

	return nil
}
