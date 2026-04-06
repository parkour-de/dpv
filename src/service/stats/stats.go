package stats

import (
	"context"
	"dpv/dpv/src/repository/graph"
	"dpv/dpv/src/repository/t"
)

type StatsResponse struct {
	ActiveUsers   int `json:"active_users"`
	ActiveClubs   int `json:"active_clubs"`
	ActiveMembers int `json:"active_members"`
}

type Service struct {
	DB *graph.Db
}

func NewService(db *graph.Db) *Service {
	return &Service{DB: db}
}

func (s *Service) GetStats(ctx context.Context) (*StatsResponse, error) {
	resp := &StatsResponse{}

	// Query active users
	queryUsers := `FOR u IN users FILTER u.membership.status == "active" COLLECT WITH COUNT INTO length RETURN length`
	cursorUsers, err := s.DB.Database.Query(ctx, queryUsers, nil)
	if err == nil {
		defer cursorUsers.Close()
		var countUsers int
		_, err = cursorUsers.ReadDocument(ctx, &countUsers)
		if err == nil {
			resp.ActiveUsers = countUsers
		}
	} else {
		return nil, t.Errorf("failed to query active users: %w", err)
	}

	// Query active clubs and their sum of members
	queryClubs := `FOR c IN clubs FILTER c.membership.status == "active" COLLECT AGGREGATE totalClubs = LENGTH(1), totalMembers = SUM(c.members) RETURN {clubs: totalClubs, members: totalMembers}`
	cursorClubs, err := s.DB.Database.Query(ctx, queryClubs, nil)
	if err == nil {
		defer cursorClubs.Close()
		var res struct {
			Clubs   int `json:"clubs"`
			Members int `json:"members"`
		}
		_, err = cursorClubs.ReadDocument(ctx, &res)
		if err == nil {
			resp.ActiveClubs = res.Clubs
			resp.ActiveMembers = res.Members
		}
	} else {
		return nil, t.Errorf("failed to query active clubs: %w", err)
	}

	return resp, nil
}
