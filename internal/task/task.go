package task

import (
	"errors"
	"time"
)

// Status represents the status of a task
type Status string

const (
	StatusPending   Status = "pending"
	StatusCompleted Status = "completed"
	StatusScheduled Status = "scheduled"
)

// Reminder represents a task reminder
type Reminder struct {
	Time         time.Time
	Acknowledged bool
}

// Task represents a task item
type Task struct {
	DueDate   *time.Time `json:"due_date,omitempty"`
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Project   string     `json:"project"`
	Phase     string     `json:"phase"`
	Status    Status     `json:"status"`
	Comments  []string   `json:"comments,omitempty"`
	Tags      []string   `json:"tags,omitempty"`
	Reminders []Reminder `json:"reminders,omitempty"`
}

// Common errors
var (
	ErrTaskNotFound  = errors.New("task not found")
	ErrInvalidTaskID = errors.New("invalid task ID")
	ErrEmptyTitle    = errors.New("task title cannot be empty")
	ErrInvalidStatus = errors.New("invalid task status")
)

// Validate validates the task fields
func (t *Task) Validate() error {
	if t.Title == "" {
		return ErrEmptyTitle
	}
	if t.Status != StatusPending && t.Status != StatusCompleted && t.Status != StatusScheduled {
		return ErrInvalidStatus
	}
	if t.Project == "" {
		return errors.New("task project cannot be empty")
	}
	if t.Phase == "" {
		return errors.New("task phase cannot be empty")
	}
	return nil
}

// Complete marks the task as completed
func (t *Task) Complete() {
	t.Status = StatusCompleted
}

// Reopen marks the task as pending
func (t *Task) Reopen() {
	t.Status = StatusPending
}
