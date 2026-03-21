package graph

import (
	"context"
	"fmt"

	"github.com/arangodb/go-driver/v2/arangodb"
)

// ProcessMembershipTransitions executes bulk AQL updates to transition 'approved' -> 'active'
// and 'cancelling' -> 'cancelled' (or 'inactive').
func (db *Db) ProcessMembershipTransitions(ctx context.Context, nowUnix int64) error {
	queries := []string{
		// Users: approved -> active
		`FOR doc IN users
		  FILTER doc.membership.status == "approved" AND doc.membership.begin_date <= @now AND doc.membership.begin_date > 0
		  UPDATE doc WITH { membership: { status: "active" } } IN users`,
		// Clubs: approved -> active
		`FOR doc IN clubs
		  FILTER doc.membership.status == "approved" AND doc.membership.begin_date <= @now AND doc.membership.begin_date > 0
		  UPDATE doc WITH { membership: { status: "active" } } IN clubs`,
		// Users: cancelling -> cancelled
		`FOR doc IN users
		  FILTER doc.membership.status == "cancelling" AND doc.membership.end_date <= @now AND doc.membership.end_date > 0
		  UPDATE doc WITH { membership: { status: "cancelled", end_date: 0, begin_date: 0 } } IN users`,
		// Clubs: cancelling -> cancelled
		`FOR doc IN clubs
		  FILTER doc.membership.status == "cancelling" AND doc.membership.end_date <= @now AND doc.membership.end_date > 0
		  UPDATE doc WITH { membership: { status: "cancelled", end_date: 0, begin_date: 0 } } IN clubs`,
	}

	bindVars := map[string]interface{}{
		"now": nowUnix,
	}

	for _, q := range queries {
		cursor, err := db.Database.Query(ctx, q, &arangodb.QueryOptions{BindVars: bindVars})
		if err != nil {
			return fmt.Errorf("failed to process membership transition: %w", err)
		}
		cursor.Close()
	}

	return nil
}
