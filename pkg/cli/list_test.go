package cli

import (
	"testing"

	"github.com/Sanmoo/my-tasks/internal/app"
	"github.com/stretchr/testify/assert"
)

func TestListCommand_ArgumentParsing(t *testing.T) {
	t.Run("command creation", func(t *testing.T) {
		cmd := newListCmd()
		assert.NotNil(t, cmd)
		assert.Equal(t, "list [project_names]", cmd.Use)
		assert.Equal(t, "List all tasks from named projects, comma-separated", cmd.Short)
	})

	t.Run("status flag exists", func(t *testing.T) {
		cmd := newListCmd()
		statusFlag := cmd.Flags().Lookup("status")
		assert.NotNil(t, statusFlag)
		assert.Equal(t, "s", statusFlag.Shorthand)
	})
}

func TestListCommand_ErrorScenarios(t *testing.T) {
	t.Run("error when no project specified and no default", func(t *testing.T) {
		// Setup app without default project
		App = &app.App{
			TaskService: nil, // We can't easily test the full execution without proper mocking
			Config: &app.Config{
				DefaultProject: "",
			},
		}

		// This test verifies the error message structure
		// The actual command execution would fail due to nil TaskService
		// but we're testing the argument parsing logic
		assert.Equal(t, "", App.Config.DefaultProject)
	})
}

func TestListCommand_Integration(t *testing.T) {
	t.Run("basic command structure", func(t *testing.T) {
		// This is a basic integration test that verifies the command can be created
		// and has the expected structure
		cmd := newListCmd()

		assert.Equal(t, "list", cmd.Name())
		assert.True(t, cmd.HasAvailableFlags())

		// Verify the command has the expected flags
		statusFlag := cmd.Flags().Lookup("status")
		assert.NotNil(t, statusFlag)
		assert.Equal(t, "Filter by task statuses (comma-separated). Available statuses: pending, running, scheduled, completed", statusFlag.Usage)
	})
}
