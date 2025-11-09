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

func (s *Service) GetProjectsByNamesOrAliases(ctx context.Context, projectNameOrAliases []string) ([]*Project, error) {
	p, err := s.repo.GetProjects(ctx, projectNameOrAliases)
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
