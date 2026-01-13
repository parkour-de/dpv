package graph

import (
	"context"
	"dpv/dpv/src/domain/entities"
	"dpv/dpv/src/repository/t"

	"github.com/arangodb/go-driver/v2/arangodb"
)

// GetCensus returns the census for a club and year.
func (db *Db) GetCensus(ctx context.Context, clubKey string, year int) (*entities.Census, error) {
	query := `
		FOR v, e IN 1..1 OUTBOUND @clubKey edges
			FILTER e.type == "census" AND e.year == @year
			RETURN v
	`
	bindVars := map[string]interface{}{
		"clubKey": "clubs/" + clubKey,
		"year":    year,
	}

	cursor, err := db.Database.Query(ctx, query, &arangodb.QueryOptions{BindVars: bindVars})
	if err != nil {
		return nil, t.Errorf("query for census failed: %w", err)
	}
	defer cursor.Close()

	if !cursor.HasMore() {
		return nil, t.Errorf("census not found")
	}

	var result entities.Census
	_, err = cursor.ReadDocument(ctx, &result)
	if err != nil {
		return nil, t.Errorf("obtaining census document failed: %w", err)
	}

	return &result, nil
}

// UpsertCensus creates or updates a census for a club and year.
func (db *Db) UpsertCensus(ctx context.Context, clubKey string, census *entities.Census) error {
	census.MemberCount = len(census.Members)

	// 1. Create/Update the Census Node
	// We do an UPSERT based on some unique criteria if possible, but Census nodes don't naturally have a unique key other than ID.
	// Actually, we can create a new node every time, or try to find one connected to the club for that year.

	// Let's try to find if one exists first.
	existing, err := db.GetCensus(ctx, clubKey, census.Year)
	if err == nil && existing != nil {
		// Update existing
		census.SetKey(existing.GetKey())
		err := db.Censuses.Update(census, ctx)
		if err != nil {
			return t.Errorf("failed to update census node: %w", err)
		}
	} else {
		// Create new
		err := db.Censuses.Create(census, ctx)
		if err != nil {
			return t.Errorf("failed to create census node: %w", err)
		}

		// Create Edge
		// We can use UPSERT logic for the edge to ensure only one exists for that year
		query := `
			UPSERT { _from: @clubKey, _to: @censusId, type: "census", year: @year }
			INSERT { _from: @clubKey, _to: @censusId, type: "census", year: @year }
			UPDATE { year: @year } IN edges
		`
		bindVars := map[string]interface{}{
			"clubKey":  "clubs/" + clubKey,
			"censusId": "censuses/" + census.GetKey(),
			"year":     census.Year,
		}
		_, err = db.Database.Query(ctx, query, &arangodb.QueryOptions{BindVars: bindVars})
		if err != nil {
			return t.Errorf("failed to create census edge: %w", err)
		}
	}

	return nil
}
