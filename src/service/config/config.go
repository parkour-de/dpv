package config

import (
	"context"
	"dpv/dpv/src/domain/entities"
	"dpv/dpv/src/repository/graph"
)

type Service struct {
	DB *graph.Db
}

func NewService(db *graph.Db) *Service {
	return &Service{
		DB: db,
	}
}

func (s *Service) GetConfig(ctx context.Context) (*entities.Config, error) {
	return s.DB.GetConfig(ctx)
}

func (s *Service) UpdateLinks(ctx context.Context, links map[string]string) (*entities.Config, error) {
	return s.DB.UpdateLinks(ctx, links)
}
