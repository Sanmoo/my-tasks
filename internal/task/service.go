package task

import (
	"context"
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
func (s *Service) CreateTask(ctx context.Context, taskTitle, projectNameOrAlias, phase string, order int) error {
	task, err := NewTask(taskTitle)
	if err != nil {
		return err
	}

	p, err := s.repo.GetProject(ctx, projectNameOrAlias)
	if err != nil {
		return err
	}

	ph, err := p.GetPhaseByName(phase)
	if err != nil {
		return err
	}

	ph.AddTask(task)

	return nil
}

func (s *Service) GetProjectByNameOrAlias(ctx context.Context, projectNameOrAlias string) (*Project, error) {
	p, err := s.repo.GetProject(ctx, projectNameOrAlias)
	if err != nil {
		return nil, err
	}

	return p, nil
}
