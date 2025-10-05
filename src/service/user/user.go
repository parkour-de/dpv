package user

import (
	"context"
	"dpv/dpv/src/domain/entities"
	"dpv/dpv/src/repository/graph"
)

type Service struct {
	DB *graph.Db
}

func NewService(db *graph.Db) *Service {
	return &Service{DB: db}
}

func (s *Service) CreateUser(ctx context.Context, user *entities.User) error {
	return s.DB.Users.Create(user, ctx)
}
