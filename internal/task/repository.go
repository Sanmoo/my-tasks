package task

import "context"

// Repository defines the interface for task storage operations
type Repository interface {
	// Create creates a new task
	Create(ctx context.Context, task *Task) error

	// GetByID retrieves a task by its ID
	GetByID(ctx context.Context, id string) (*Task, error)

	// List retrieves all tasks with optional filtering
	List(ctx context.Context, filter Filter) ([]*Task, error)

	// Update updates an existing task
	Update(ctx context.Context, task *Task) error

	// Delete deletes a task by its ID
	Delete(ctx context.Context, id string) error
}

// Filter represents filtering options for listing tasks
type Filter struct {
	Status   *Status
	Tags     []string
	Priority *int
}
