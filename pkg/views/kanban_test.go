package views

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

// testKanban wraps the original Kanban but captures output
type testKanban struct {
	output *bytes.Buffer
	Kanban
}

func (tk *testKanban) Render() {
	// Capture the output instead of printing to stdout
	fmt.Fprintf(tk.output, "========================== %s ==========================\n\n", tk.ProjectName)
	for _, column := range tk.Columns {
		for _, task := range column.Tasks {
			fmt.Fprintf(tk.output, "[%s] %s\n", column.Name, task)
		}
	}
	fmt.Fprintf(tk.output, "\n========================== %s ==========================\n", tk.ProjectName)
}

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
				ProjectName: "Simple Project",
				Columns: []Column{
					{
						Name:  "📋 Backlog",
						Tasks: []string{"Task 1", "Task 2", "Task 3"},
					},
				},
			},
		},
		{
			name: "multiple columns",
			kanban: Kanban{
				ProjectName: "Complex Project",
				Columns: []Column{
					{
						Name:  "📋 Backlog",
						Tasks: []string{"Task 1", "Task 2"},
					},
					{
						Name:  "🏃 In Progress",
						Tasks: []string{"Task 3"},
					},
					{
						Name:  "✅ Done",
						Tasks: []string{"Task 4", "Task 5", "Task 6"},
					},
				},
			},
		},
		{
			name: "column with no tasks",
			kanban: Kanban{
				ProjectName: "Mixed Project",
				Columns: []Column{
					{
						Name:  "📋 Backlog",
						Tasks: []string{"Task 1"},
					},
					{
						Name:  "🏃 In Progress",
						Tasks: []string{},
					},
					{
						Name:  "✅ Done",
						Tasks: []string{"Task 2"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			tk := testKanban{
				Kanban: tt.kanban,
				output: &buf,
			}
			tk.Render()
			snaps.MatchSnapshot(t, buf.String())
		})
	}
}
