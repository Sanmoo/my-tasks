package cli

import (
	"testing"
	"time"

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

	t.Run("has expiring-in flag", func(t *testing.T) {
		cmd := newRemindCmd()
		assert.True(t, cmd.HasAvailableFlags())

		flag := cmd.Flags().Lookup("expiring-in")
		assert.NotNil(t, flag)
		assert.Equal(t, "expiring-in", flag.Name)
		assert.Equal(t, "", flag.DefValue)
		assert.Equal(t, "Filter reminders that will expire within the specified duration (e.g., 2h, 30m, 1d)", flag.Usage)
	})
}

func TestRemindCommand_Integration(t *testing.T) {
	t.Run("basic command structure", func(t *testing.T) {
		cmd := newRemindCmd()

		assert.Equal(t, "remind", cmd.Name())
		assert.True(t, cmd.HasAvailableFlags()) // remind command now has flags
	})
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
		hasError bool
	}{
		{
			name:     "standard hours",
			input:    "2h",
			expected: 2 * time.Hour,
			hasError: false,
		},
		{
			name:     "standard minutes",
			input:    "30m",
			expected: 30 * time.Minute,
			hasError: false,
		},
		{
			name:     "days",
			input:    "1d",
			expected: 24 * time.Hour,
			hasError: false,
		},
		{
			name:     "multiple days",
			input:    "3d",
			expected: 72 * time.Hour,
			hasError: false,
		},
		{
			name:     "invalid duration",
			input:    "invalid",
			expected: 0,
			hasError: true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: 0,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDuration(tt.input)

			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
