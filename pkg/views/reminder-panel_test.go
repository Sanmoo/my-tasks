package views

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/gkampitakis/go-snaps/snaps"
)

// testReminderPanel wraps the original ReminderPanel but captures output
type testReminderPanel struct {
	ReminderPanel
	output *bytes.Buffer
}

func (trp *testReminderPanel) Render() {
	if len(trp.Reminders) == 0 {
		trp.output.WriteString("No reminders are overdue")
		return
	}

	trp.output.WriteString("============= !You Have Overdue Reminders. Please check them =============\n\n")

	// Sort projects for deterministic output
	projects := make([]string, 0, len(trp.Reminders))
	for project := range trp.Reminders {
		projects = append(projects, project)
	}
	// Sort projects alphabetically
	for i := 0; i < len(projects); i++ {
		for j := i + 1; j < len(projects); j++ {
			if projects[i] > projects[j] {
				projects[i], projects[j] = projects[j], projects[i]
			}
		}
	}

	for _, project := range projects {
		fmt.Fprintf(trp.output, "PROJECT: %s\n", project)
		for _, rem := range trp.Reminders[project] {
			fmt.Fprintf(trp.output, "- %s\n", rem.Label)
		}
	}
	trp.output.WriteString("\n==========================================================================\n")
}

func TestReminderPanel_Render(t *testing.T) {
	tests := []struct {
		reminder ReminderPanel
		name     string
	}{
		{
			name:     "empty reminder panel",
			reminder: ReminderPanel{Reminders: make(map[string][]Reminder)},
		},
		{
			name: "single project with single reminder",
			reminder: func() ReminderPanel {
				rp := NewReminderPanel()
				rp.AddReminder(Reminder{
					Project: "Project A",
					Label:   "Complete task documentation",
					Due:     time.Now().Add(-24 * time.Hour),
				})
				return *rp
			}(),
		},
		{
			name: "single project with multiple reminders",
			reminder: func() ReminderPanel {
				rp := NewReminderPanel()
				rp.AddReminder(Reminder{
					Project: "Project A",
					Label:   "Complete task documentation",
					Due:     time.Now().Add(-24 * time.Hour),
				})
				rp.AddReminder(Reminder{
					Project: "Project A",
					Label:   "Review pull request #123",
					Due:     time.Now().Add(-12 * time.Hour),
				})
				rp.AddReminder(Reminder{
					Project: "Project A",
					Label:   "Update dependencies",
					Due:     time.Now().Add(-6 * time.Hour),
				})
				return *rp
			}(),
		},
		{
			name: "multiple projects with reminders",
			reminder: func() ReminderPanel {
				rp := NewReminderPanel()
				rp.AddReminder(Reminder{
					Project: "Project A",
					Label:   "Complete task documentation",
					Due:     time.Now().Add(-24 * time.Hour),
				})
				rp.AddReminder(Reminder{
					Project: "Project B",
					Label:   "Fix critical bug in authentication",
					Due:     time.Now().Add(-48 * time.Hour),
				})
				rp.AddReminder(Reminder{
					Project: "Project B",
					Label:   "Update API documentation",
					Due:     time.Now().Add(-36 * time.Hour),
				})
				rp.AddReminder(Reminder{
					Project: "Project C",
					Label:   "Prepare for team meeting",
					Due:     time.Now().Add(-2 * time.Hour),
				})
				return *rp
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Call the actual Render method
			tt.reminder.Render()

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

func TestReminderPanel_AddReminder(t *testing.T) {
	t.Run("add reminder to new project", func(t *testing.T) {
		rp := NewReminderPanel()
		rp.AddReminder(Reminder{
			Project: "New Project",
			Label:   "Test reminder",
			Due:     time.Now(),
		})

		if len(rp.Reminders["New Project"]) != 1 {
			t.Errorf("Expected 1 reminder, got %d", len(rp.Reminders["New Project"]))
		}
		if rp.Reminders["New Project"][0].Label != "Test reminder" {
			t.Errorf("Expected label 'Test reminder', got '%s'", rp.Reminders["New Project"][0].Label)
		}
	})

	t.Run("add multiple reminders to same project", func(t *testing.T) {
		rp := NewReminderPanel()
		rp.AddReminder(Reminder{
			Project: "Project A",
			Label:   "First reminder",
			Due:     time.Now(),
		})
		rp.AddReminder(Reminder{
			Project: "Project A",
			Label:   "Second reminder",
			Due:     time.Now(),
		})

		if len(rp.Reminders["Project A"]) != 2 {
			t.Errorf("Expected 2 reminders, got %d", len(rp.Reminders["Project A"]))
		}
	})
}
