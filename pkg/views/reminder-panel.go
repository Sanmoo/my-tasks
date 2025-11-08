package views

import (
	"fmt"
	"time"
)

type ReminderPanel struct {
	Reminders []Reminder
}

type Reminder struct {
	Due     time.Time
	Label   string
	Project string
}

func (rp *ReminderPanel) Render() {
	if len(rp.Reminders) == 0 {
		fmt.Printf("No reminders are overdue")
		return
	}

	fmt.Printf("============= !You Have Overdue Reminders. Please check them =============\n\n")
	for _, rem := range rp.Reminders {
		fmt.Printf("- %s (PROJECT: %s)\n", rem.Label, rem.Project)
	}
	fmt.Printf("\n==========================================================================\n")
}
