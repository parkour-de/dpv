package graph

import (
	"context"
	"dpv/dpv/src/domain/entities"
	"fmt"

	"github.com/arangodb/go-driver/v2/arangodb"
	"github.com/arangodb/go-driver/v2/arangodb/shared"
)

type TransitionResult struct {
	ActivatedUsers []entities.User
	ActivatedClubs []entities.Club
	CancelledUsers []entities.User
	CancelledClubs []entities.Club
}

// ProcessMembershipTransitions executes bulk AQL updates to transition 'approved' -> 'active'
// and 'cancelling' -> 'cancelled' (or 'inactive').
func (db *Db) ProcessMembershipTransitions(ctx context.Context, nowUnix int64) (*TransitionResult, error) {
	result := &TransitionResult{
		ActivatedUsers: make([]entities.User, 0),
		ActivatedClubs: make([]entities.Club, 0),
		CancelledUsers: make([]entities.User, 0),
		CancelledClubs: make([]entities.Club, 0),
	}

	queries := []struct {
		Query string
		Type  string // "users_activated", "clubs_activated", etc.
	}{
		{
			Query: `FOR doc IN users
		  FILTER doc.membership.status == "approved" AND doc.membership.begin_date <= @now AND doc.membership.begin_date > 0
		  UPDATE doc WITH { membership: { status: "active" } } IN users RETURN NEW`,
			Type: "users_activated",
		},
		{
			Query: `FOR doc IN clubs
		  FILTER doc.membership.status == "approved" AND doc.membership.begin_date <= @now AND doc.membership.begin_date > 0
		  UPDATE doc WITH { membership: { status: "active" } } IN clubs RETURN NEW`,
			Type: "clubs_activated",
		},
		{
			Query: `FOR doc IN users
		  FILTER doc.membership.status == "cancelling" AND doc.membership.end_date <= @now AND doc.membership.end_date > 0
		  UPDATE doc WITH { membership: { status: "cancelled", end_date: 0, begin_date: 0 } } IN users RETURN NEW`,
			Type: "users_cancelled",
		},
		{
			Query: `FOR doc IN clubs
		  FILTER doc.membership.status == "cancelling" AND doc.membership.end_date <= @now AND doc.membership.end_date > 0
		  UPDATE doc WITH { membership: { status: "cancelled", end_date: 0, begin_date: 0 } } IN clubs RETURN NEW`,
			Type: "clubs_cancelled",
		},
	}

	bindVars := map[string]interface{}{
		"now": nowUnix,
	}

	for _, q := range queries {
		cursor, err := db.Database.Query(ctx, q.Query, &arangodb.QueryOptions{BindVars: bindVars})
		if err != nil {
			return nil, fmt.Errorf("failed to process membership transition: %w", err)
		}
		
		for {
			var err2 error
			if q.Type == "users_activated" || q.Type == "users_cancelled" {
				var user entities.User
				_, err2 = cursor.ReadDocument(ctx, &user)
				if shared.IsNoMoreDocuments(err2) {
					break // This correctly breaks the for loop because it's not inside a switch
				}
				if err2 != nil {
					return nil, fmt.Errorf("failed to decode user: %w", err2)
				}
				if q.Type == "users_activated" {
					result.ActivatedUsers = append(result.ActivatedUsers, user)
				} else {
					result.CancelledUsers = append(result.CancelledUsers, user)
				}
			} else if q.Type == "clubs_activated" || q.Type == "clubs_cancelled" {
				var club entities.Club
				_, err2 = cursor.ReadDocument(ctx, &club)
				if shared.IsNoMoreDocuments(err2) {
					break // This correctly breaks the for loop because it's not inside a switch
				}
				if err2 != nil {
					return nil, fmt.Errorf("failed to decode club: %w", err2)
				}
				// We need to fetch the Vorstand
				fullClub, err3 := db.GetClubByKey(ctx, club.Key)
				if err3 == nil && fullClub != nil {
					club = *fullClub
				}

				if q.Type == "clubs_activated" {
					result.ActivatedClubs = append(result.ActivatedClubs, club)
				} else {
					result.CancelledClubs = append(result.CancelledClubs, club)
				}
			}
		}
		
		cursor.Close()
	}

	return result, nil
}
