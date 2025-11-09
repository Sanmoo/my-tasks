package task

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
)

// Status represents the status of a task
type Status string

const (
	StatusPending   Status = "pending"
	StatusCompleted Status = "completed"
	StatusScheduled Status = "scheduled"
	StatusRunning   Status = "running"
)

type Project struct {
	name   string
	phases []*Phase
}

func (p *Project) GetPhases() []*Phase {
	return p.phases
}

func (p *Project) GetName() string {
	return p.name
}

type Phase struct {
	name             string
	associatedStatus Status
	projectName      string
	tasks            []*Task
}

func (ph *Phase) GetName() string {
	return ph.name
}

type Task struct {
	dueDate   *time.Time
	title     string
	comments  []string
	tags      []string
	reminders []*Reminder
}

// Reminder represents a task Reminder
type Reminder struct {
	Label        string
	Time         time.Time
	TaskTitle    string
	Acknowledged bool
}

func NewProject(name string) (*Project, error) {
	if name == "" {
		return nil, errors.New("a project Name must not be empty")
	}

	// At least a Phase for pending and another one for concluded
	return &Project{
		name:   name,
		phases: []*Phase{},
	}, nil
}

func (p *Project) AddPhase(ph *Phase) {
	p.phases = append(p.phases, ph)
}

func NewPhase(name, projectName string) (*Phase, error) {
	if name == "" {
		return nil, errors.New("a project Name must not be empty")
	}

	var associatedStatus Status
	switch {
	case strings.Contains(name, "🏃"):
		associatedStatus = StatusRunning
	case strings.Contains(name, "🗓️"):
		associatedStatus = StatusScheduled
	case strings.Contains(name, "✅"):
		associatedStatus = StatusCompleted
	default:
		associatedStatus = StatusPending
	}

	return &Phase{
		name:             name,
		associatedStatus: associatedStatus,
		projectName:      projectName,
	}, nil
}

func NewTask(title string) (*Task, error) {
	if title == "" {
		return nil, errors.New("a project name must not be empty")
	}

	return &Task{
		title: title,
	}, nil
}

func (task *Task) AddComment(comment string) error {
	if comment == "" {
		return errors.New("a comment must not be empty")
	}

	task.comments = append(task.comments, comment)
	return nil
}

func (task *Task) AddReminder(reminder *Reminder) {
	task.reminders = append(task.reminders, reminder)
}

func (task *Task) SetDueDate(dueDate *time.Time) error {
	if task.dueDate != nil {
		return fmt.Errorf("can't set due date %s for task %s since it is already set to %s", dueDate, task.title, task.dueDate)
	}

	task.dueDate = dueDate
	return nil
}

func (task *Task) AddTag(tag string) error {
	if slices.Contains(task.tags, tag) {
		return fmt.Errorf("can't declare the same tag %s twice for task %s", tag, task.title)
	}

	task.tags = append(task.tags, tag)
	return nil
}

func (ph *Phase) AddTask(task *Task) {
	ph.tasks = append(ph.tasks, task)
}

func (p *Project) GetPhaseByName(name string) (*Phase, error) {
	for _, phase := range p.phases {
		if phase.name == name {
			return phase, nil
		}
	}

	return nil, fmt.Errorf("phase %s not found in project with name %s", name, p.name)
}

func (ph *Phase) GetTasks() []*Task {
	return ph.tasks
}

func (ph *Phase) GetTaskTitles() []string {
	result := []string{}
	for _, task := range ph.GetTasks() {
		result = append(result, task.GetTitle())
	}
	return result
}

func (task *Task) GetTags() []string {
	return task.tags
}

func (task *Task) GetTitle() string {
	return task.title
}

func (task *Task) GetOverdueReminders() []*Reminder {
	result := []*Reminder{}

	for _, rem := range task.reminders {
		if !rem.Acknowledged && rem.Time.Before(time.Now()) {
			rem.TaskTitle = task.title
			result = append(result, rem)
		}
	}

	return result
}

func (ph *Phase) GetOverdueReminders() []*Reminder {
	result := []*Reminder{}

	for _, task := range ph.GetTasks() {
		result = append(result, task.GetOverdueReminders()...)
	}

	return result
}

func (p *Project) GetWarnings() []string {
	result := []string{}

	for _, ph := range p.GetPhases() {
		result = append(result, ph.GetWarnings()...)
	}

	return result
}

func (ph *Phase) GetWarnings() []string {
	result := []string{}

	if ph.associatedStatus == StatusScheduled {
		for _, task := range ph.tasks {
			if len(task.GetActiveReminders()) == 0 {
				result = append(
					result,
					fmt.Sprintf(
						"This task [%s] - %s is in phase %s, which is of kind 'scheduled', however it has no reminders active",
						ph.projectName,
						task.title,
						ph.name,
					),
				)
			}
		}
	}

	return result
}

func (ph *Phase) GetAssociatedStatus() Status {
	return ph.associatedStatus
}

func (p *Project) GetOverdueReminders() []*Reminder {
	result := []*Reminder{}

	for _, phase := range p.GetPhases() {
		result = append(result, phase.GetOverdueReminders()...)
	}

	return result
}

func (task *Task) GetActiveReminders() []*Reminder {
	result := []*Reminder{}

	for _, rem := range task.reminders {
		if !rem.Acknowledged {
			result = append(result, rem)
		}
	}

	return result
}
