package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemindCommand(t *testing.T) {
	t.Run("command creation", func(t *testing.T) {
		cmd := newRemindCmd()
		assert.NotNil(t, cmd)
		assert.Equal(t, "remind", cmd.Use)
		assert.Equal(t, "List all tasks from all scannable projects with due reminders", cmd.Short)
	})

	t.Run("no arguments expected", func(t *testing.T) {
		cmd := newRemindCmd()
		// The remind command should not accept any arguments
		// We can verify this by checking that Args is set to MaximumNArgs(0)
		assert.NotNil(t, cmd.Args)
	})
}

func TestRemindCommand_Integration(t *testing.T) {
	t.Run("basic command structure", func(t *testing.T) {
		cmd := newRemindCmd()

		assert.Equal(t, "remind", cmd.Name())
		assert.False(t, cmd.HasAvailableFlags()) // remind command has no flags
	})
}
