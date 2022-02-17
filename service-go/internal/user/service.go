package user

import (
	"context"

	"github.com/google/uuid"
)

type Repo interface {
	GetOrCreate(ctx context.Context, request User) (User, error)
}

type CRUD struct {
	repo Repo
}

func NewCRUD(repo Repo) *CRUD {
	return &CRUD{
		repo: repo,
	}
}

func (s *CRUD) GetOrCreate(ctx context.Context, request User) (User, error) {
	request.ID = uuid.New()
	for i := range request.Idents {
		request.Idents[i].ID = uuid.New()
	}

	return s.repo.GetOrCreate(ctx, request)
}
