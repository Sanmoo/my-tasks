package task

import "context"

type Repository interface {
	GetProjects(ctx context.Context, namesOrAliases []string) ([]*Project, error)
	GetAllProjects(ctx context.Context) ([]*Project, error)
}
