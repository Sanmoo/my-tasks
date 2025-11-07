package task

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Service provides business logic for task operations
type Service struct {
	repo Repository
}

// NewService creates a new task service
func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// CreateTask creates a new task with validation
func (s *Service) CreateTask(ctx context.Context, title, project, phase string, priority int, tags []string, comments []string) (*Task, error) {
	task := &Task{
		ID:       uuid.New().String(),
		Title:    title,
		Project:  project,
		Phase:    phase,
		Status:   StatusPending,
		Tags:     tags,
		Comments: comments,
	}

	if err := task.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, task); err != nil {
		return nil, err
	}

	return task, nil
}

// GetTask retrieves a task by ID
func (s *Service) GetTask(ctx context.Context, id string) (*Task, error) {
	if id == "" {
		return nil, ErrInvalidTaskID
	}
	return s.repo.GetByID(ctx, id)
}

// ListTasks retrieves tasks with optional filtering
func (s *Service) ListTasks(ctx context.Context, filter Filter) ([]*Task, error) {
	return s.repo.List(ctx, filter)
}

// CompleteTask marks a task as completed
func (s *Service) CompleteTask(ctx context.Context, id string) error {
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	task.Complete()
	return s.repo.Update(ctx, task)
}

// ReopenTask marks a task as pending
func (s *Service) ReopenTask(ctx context.Context, id string) error {
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	task.Reopen()
	return s.repo.Update(ctx, task)
}

// UpdateTask updates a task's details
func (s *Service) UpdateTask(ctx context.Context, id, title, project, phase string, priority int, tags, comments []string, dueDate *time.Time) error {
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if title != "" {
		task.Title = title
	}
	if project != "" {
		task.Project = project
	}
	if phase != "" {
		task.Phase = phase
	}
	if tags != nil {
		task.Tags = tags
	}
	if comments != nil {
		task.Comments = comments
	}
	if dueDate != nil {
		task.DueDate = dueDate
	}

	if err := task.Validate(); err != nil {
		return err
	}

	return s.repo.Update(ctx, task)
}

// DeleteTask deletes a task
func (s *Service) DeleteTask(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidTaskID
	}
	return s.repo.Delete(ctx, id)
}
