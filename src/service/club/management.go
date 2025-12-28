package club

import (
	"context"
	"dpv/dpv/src/domain/entities"
	"dpv/dpv/src/repository/graph"
)

// GetAllClubs retrieves all clubs with optional filtering and pagination.
func (s *Service) GetAllClubs(ctx context.Context, options graph.ClubQueryOptions) ([]entities.Club, error) {
	return s.DB.GetClubs(ctx, options)
}
