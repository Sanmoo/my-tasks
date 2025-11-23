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
