package views

import (
	"fmt"
	"sort"
	"time"
)

type ReminderPanel struct {
	Reminders map[string][]Reminder
}

type Reminder struct {
	Due     time.Time
	Label   string
	Project string
}

func NewReminderPanel() *ReminderPanel {
	return &ReminderPanel{
		Reminders: make(map[string][]Reminder),
	}
}

func (rp *ReminderPanel) Render() {
	if len(rp.Reminders) == 0 {
		fmt.Printf("No reminders are overdue")
		return
	}

	fmt.Printf("============= !You Have Overdue Reminders. Please check them =============\n\n")

	// Sort projects for deterministic output
	projects := make([]string, 0, len(rp.Reminders))
	for project := range rp.Reminders {
		projects = append(projects, project)
	}
	sort.Strings(projects)

	for _, project := range projects {
		fmt.Printf("PROJECT: %s\n", project)
		for _, rem := range rp.Reminders[project] {
			fmt.Printf("- %s\n", rem.Label)
		}
	}
	fmt.Printf("\n==========================================================================\n")
}

func (rp *ReminderPanel) AddReminder(r Reminder) {
	if rp.Reminders[r.Project] == nil {
		rp.Reminders[r.Project] = make([]Reminder, 0)
	}
	rp.Reminders[r.Project] = append(rp.Reminders[r.Project], r)
}
