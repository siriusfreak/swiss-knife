package savedjql

import (
	"context"
	"github.com/siriusfreak/swiss-knife/internal/app/saved-jql/repository"
)

type SavedJQL = repository.SavedJQL
type Repository interface {
	CreateTables(ctx context.Context) error
	SaveJQL(ctx context.Context, jql *SavedJQL) error
	GetAllJQL(ctx context.Context) ([]*SavedJQL, error)
	DeleteJQL(ctx context.Context, id string) error
}

type Service struct {
	repo Repository
}

func New(ctx context.Context, repo Repository) (*Service, error) {
	err := repo.CreateTables(ctx)
	if err != nil {
		return nil, err
	}
	return &Service{repo: repo}, nil
}

func (s *Service) CreateTables(ctx context.Context) error {
	return s.repo.CreateTables(ctx)
}

func (s *Service) SaveJQL(ctx context.Context, jql *SavedJQL) error {
	return s.repo.SaveJQL(ctx, jql)
}

func (s *Service) GetAllJQL(ctx context.Context) ([]*SavedJQL, error) {
	return s.repo.GetAllJQL(ctx)
}

func (s *Service) DeleteJQL(ctx context.Context, id string) error {
	return s.repo.DeleteJQL(ctx, id)
}
