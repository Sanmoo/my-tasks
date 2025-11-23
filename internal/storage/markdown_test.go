package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarkdownStorage_GetProjects(t *testing.T) {
	// Create a temporary test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.md")

	content := `# Test Project
## 🏃 In Progress
* Task 1
  * @remind (25-01-01 10:00:00)
  * @tags urgent important
  * Regular comment
* Task 2
    * @due (25-01-02)
## ✅ Done
* Completed Task
  * @reminded (25-01-01 09:00:00)
`
	err := os.WriteFile(testFile, []byte(content), 0o644)
	require.NoError(t, err)

	storage, err := NewMarkdownStorage([]string{testFile}, map[string]string{"test": "Test Project"}, "UTC")
	require.NoError(t, err)

	t.Run("Get existing project by name", func(t *testing.T) {
		projects, err := storage.GetProjects(context.Background(), []string{"Test Project"})
		require.NoError(t, err)
		require.Len(t, projects, 1)

		project := projects[0]
		assert.Equal(t, "Test Project", project.GetName())
		assert.Len(t, project.GetPhases(), 2)
	})

	t.Run("Get existing project by alias", func(t *testing.T) {
		projects, err := storage.GetProjects(context.Background(), []string{"test"})
		require.NoError(t, err)
		require.Len(t, projects, 1)
		assert.Equal(t, "Test Project", projects[0].GetName())
	})

	t.Run("Get non-existent project", func(t *testing.T) {
		projects, err := storage.GetProjects(context.Background(), []string{"NonExistent"})
		assert.Error(t, err)
		assert.Nil(t, projects)
	})

	t.Run("Get all projects", func(t *testing.T) {
		projects, err := storage.GetAllProjects(context.Background())
		require.NoError(t, err)
		require.Len(t, projects, 1)
		assert.Equal(t, "Test Project", projects[0].GetName())
	})
}

func TestMarkdownStorage_ParseDirectives(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "directives.md")

	content := `# Directives Test
## 📋 Backlog
* Test Task
  * @remind (25-01-01 10:00:00)
  * @reminded (25-01-01 09:00:00)
  * @tags tag1 tag2
  * @due (25-01-03)
  * Regular comment line
    * Sub-subbullet @remind (25-01-02 11:00:00)
`
	err := os.WriteFile(testFile, []byte(content), 0o644)
	require.NoError(t, err)

	storage, err := NewMarkdownStorage([]string{testFile}, nil, "UTC")
	require.NoError(t, err)

	projects, err := storage.GetAllProjects(context.Background())
	require.NoError(t, err)
	require.Len(t, projects, 1)

	project := projects[0]
	phases := project.GetPhases()
	require.Len(t, phases, 1)

	phase := phases[0]
	tasks := phase.GetTasks()
	require.Len(t, tasks, 1)

	task := tasks[0]
	assert.Equal(t, "Test Task", task.GetTitle())

	// Check reminders
	activeReminders := task.GetActiveReminders()
	require.Len(t, activeReminders, 1) // Only @remind (line 75) - @reminded creates acknowledged reminder

	// Check acknowledged reminder
	// allReminders := task.GetActiveReminders()
	// We need to find a better way to test this, but for now, let's check the count

	// Check tags
	tags := task.GetTags()
	assert.ElementsMatch(t, []string{"tag1", "tag2"}, tags)

	// Check due date
	// Note: Due to the task model not exposing due date directly, we might need to adjust
	// This test may need to be updated based on actual implementation

	// Check comments
	// Similarly, comments may not be directly exposed in the current model
}

func TestMarkdownStorage_EdgeCases(t *testing.T) {
	t.Run("Empty file", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "empty.md")
		err := os.WriteFile(testFile, []byte(""), 0o644)
		require.NoError(t, err)

		storage, err := NewMarkdownStorage([]string{testFile}, nil, "UTC")
		require.NoError(t, err)

		projects, err := storage.GetAllProjects(context.Background())
		require.NoError(t, err)
		assert.Empty(t, projects)
	})

	t.Run("Task without phase", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "invalid.md")
		content := `# Test Project
* Task without phase
`
		err := os.WriteFile(testFile, []byte(content), 0o644)
		require.NoError(t, err)

		storage, err := NewMarkdownStorage([]string{testFile}, nil, "UTC")
		require.NoError(t, err)

		_, err = storage.GetAllProjects(context.Background())
		assert.Error(t, err)
	})

	t.Run("Phase without project", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "invalid2.md")
		content := `## Phase without project
`
		err := os.WriteFile(testFile, []byte(content), 0o644)
		require.NoError(t, err)

		storage, err := NewMarkdownStorage([]string{testFile}, nil, "UTC")
		require.NoError(t, err)

		_, err = storage.GetAllProjects(context.Background())
		assert.Error(t, err)
	})
}

func TestParseDate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Time
		wantErr  bool
	}{
		{
			name:     "Valid date time with seconds",
			input:    "25-01-01 10:30:45 UTC",
			expected: time.Date(2025, 1, 1, 10, 30, 45, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "Valid date time without seconds",
			input:    "25-01-01 10:30 UTC",
			expected: time.Date(2025, 1, 1, 10, 30, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "Valid date only",
			input:    "25-01-01 UTC",
			expected: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:    "Invalid format",
			input:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDate(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestExtractDateFromParenthesis(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Valid parenthesis",
			input:    "@remind (25-01-01 10:00:00)",
			expected: "25-01-01 10:00:00",
		},
		{
			name:     "No parenthesis",
			input:    "@remind",
			expected: "",
		},
		{
			name:     "Empty parenthesis",
			input:    "@remind ()",
			expected: "",
		},
		{
			name:     "Multiple parenthesis",
			input:    "@remind (25-01-01) (extra)",
			expected: "25-01-01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractDateFromParenthesis(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
