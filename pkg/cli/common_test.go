package cli

import (
	"testing"

	"github.com/Sanmoo/my-tasks/internal/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderWarningsIfAny(t *testing.T) {
	t.Run("no warnings", func(t *testing.T) {
		projects := []*task.Project{}
		// This should not panic and should complete without output
		// We're testing that the function handles empty warnings gracefully
		assert.NotPanics(t, func() {
			RenderWarningsIfAny(projects)
		})
	})

	t.Run("with warnings", func(t *testing.T) {
		// Create a project with warnings
		project, err := task.NewProject("Test Project")
		require.NoError(t, err)

		phase, err := task.NewPhase("🗓️ Scheduled", "Test Project")
		require.NoError(t, err)

		taskItem, err := task.NewTask("Scheduled Task")
		require.NoError(t, err)

		phase.AddTask(taskItem)
		project.AddPhase(phase)

		projects := []*task.Project{project}

		// This should not panic and should process warnings
		// We're testing that the function handles warnings without crashing
		assert.NotPanics(t, func() {
			RenderWarningsIfAny(projects)
		})

		// Verify that the project actually has warnings
		warnings := project.GetWarnings()
		assert.Len(t, warnings, 1)
		assert.Contains(t, warnings[0], "has no reminders active")
	})
}
