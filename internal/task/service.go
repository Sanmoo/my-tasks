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

func (s *Service) GetProjectByNameOrAlias(ctx context.Context, projectNameOrAlias string) (*Project, error) {
	p, err := s.repo.GetProject(ctx, projectNameOrAlias)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (s *Service) GetAllProjects(ctx context.Context) ([]*Project, error) {
	p, err := s.repo.GetAllProjects(ctx)
	if err != nil {
		return nil, err
	}

	return p, nil
}
