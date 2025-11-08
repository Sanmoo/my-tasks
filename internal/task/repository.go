package task

import "context"

type Repository interface {
	GetProject(ctx context.Context, nameOrAlias string) (*Project, error)
	GetAllProjects(ctx context.Context) ([]*Project, error)
}
