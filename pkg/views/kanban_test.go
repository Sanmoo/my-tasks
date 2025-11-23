package views

import (
	"bytes"
	"os"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestKanban_Render(t *testing.T) {
	tests := []struct {
		name   string
		kanban Kanban
	}{
		{
			name: "empty kanban",
			kanban: Kanban{
				ProjectName: "Empty Project",
				Columns:     []Column{},
			},
		},
		{
			name: "single column with tasks",
			kanban: Kanban{
				ProjectName: "Project A",
				Columns: []Column{
					{
						Name: "TODO",
						Tasks: []string{
							"Task 1",
							"Task 2",
							"Task 3",
						},
					},
				},
			},
		},
		{
			name: "multiple columns with tasks",
			kanban: Kanban{
				ProjectName: "Project B",
				Columns: []Column{
					{
						Name: "TODO",
						Tasks: []string{
							"Design database schema",
							"Write API documentation",
						},
					},
					{
						Name: "IN PROGRESS",
						Tasks: []string{
							"Implement authentication",
						},
					},
					{
						Name: "DONE",
						Tasks: []string{
							"Setup project structure",
							"Configure CI/CD pipeline",
						},
					},
				},
			},
		},
		{
			name: "column with empty tasks",
			kanban: Kanban{
				ProjectName: "Project C",
				Columns: []Column{
					{
						Name:  "TODO",
						Tasks: []string{},
					},
					{
						Name: "IN PROGRESS",
						Tasks: []string{
							"Only task in progress",
						},
					},
				},
			},
		},
		{
			name: "columns with emoji status indicators",
			kanban: Kanban{
				ProjectName: "Project D",
				Columns: []Column{
					{
						Name: "🏃 Running",
						Tasks: []string{
							"Task in progress",
						},
					},
					{
						Name: "✅ Completed",
						Tasks: []string{
							"Finished task",
						},
					},
					{
						Name: "🗓️ Scheduled",
						Tasks: []string{
							"Future task",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Call the actual Render method
			tt.kanban.Render()

			// Restore stdout and read captured output
			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			// Use snapshot testing for the actual output
			snaps.MatchSnapshot(t, output)
		})
	}
}

func TestKanban_HelperMethods(t *testing.T) {
	t.Run("formatTask", func(t *testing.T) {
		kanban := Kanban{}

		result := kanban.formatTask("Test task")
		expected := "• Test task"
		if result != expected {
			t.Errorf("formatTask returned %q, expected %q", result, expected)
		}
	})

	t.Run("renderColumn execution without panic", func(t *testing.T) {
		kanban := Kanban{}

		// Test with empty tasks
		column1 := Column{
			Name:  "Empty Column",
			Tasks: []string{},
		}

		// Test with tasks
		column2 := Column{
			Name:  "Test Column",
			Tasks: []string{"Task 1", "Task 2"},
		}

		// Test that renderColumn executes without panicking
		testCases := []struct {
			name   string
			column Column
		}{
			{"empty tasks", column1},
			{"with tasks", column2},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("renderColumn panicked: %v", r)
					}
				}()

				// Capture stdout to prevent output during tests
				oldStdout := os.Stdout
				_, w, _ := os.Pipe()
				os.Stdout = w

				kanban.renderColumn(tc.column)

				w.Close()
				os.Stdout = oldStdout
			})
		}
	})

	t.Run("renderHorizontal execution without panic", func(t *testing.T) {
		kanban := Kanban{
			ProjectName: "Test Project",
			Columns: []Column{
				{
					Name:  "Column 1",
					Tasks: []string{"Task 1", "Task 2"},
				},
				{
					Name:  "Column 2",
					Tasks: []string{"Task 3"},
				},
			},
		}

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("renderHorizontal panicked: %v", r)
			}
		}()

		// Capture stdout to prevent output during tests
		oldStdout := os.Stdout
		_, w, _ := os.Pipe()
		os.Stdout = w

		kanban.renderHorizontal(200) // Wide terminal to force horizontal layout

		w.Close()
		os.Stdout = oldStdout
	})
}
