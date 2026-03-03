package graph

import (
	"context"
	"dpv/dpv/src/repository/t"

	"github.com/arangodb/go-driver/v2/arangodb"
)

func (db *Db) GetNextSequence(ctx context.Context, sequenceKey string) (int, error) {
	// UPSERT the config singleton to ensure it exists, then atomic increment the specific sequenceKey inside the `Sequences` map.
	// In ArangoDB AQL:
	// UPSERT { _key: "config" }
	// INSERT { _key: "config", sequences: { [sequenceKey]: 1 } }
	// UPDATE { sequences: MERGE(OLD.sequences, { [sequenceKey]: (OLD.sequences[sequenceKey] || 0) + 1 }) }
	// IN configs
	// RETURN NEW.sequences[sequenceKey]

	query := `
		UPSERT { _key: "config" }
		INSERT { _key: "config", sequences: { @sequenceKey: 1 } }
		UPDATE { sequences: MERGE(OLD.sequences || {}, { [@sequenceKey]: (OLD.sequences[@sequenceKey] || 0) + 1 }) }
		IN @@collection
		RETURN NEW.sequences[@sequenceKey]
	`
	bindVars := map[string]interface{}{
		"@collection": "configs",
		"sequenceKey": sequenceKey,
	}

	cursor, err := db.Database.Query(ctx, query, &arangodb.QueryOptions{BindVars: bindVars})
	if err != nil {
		return 0, t.Errorf("failed to query next sequence: %w", err)
	}
	defer cursor.Close()

	var nextValue int
	if _, err := cursor.ReadDocument(ctx, &nextValue); err != nil {
		return 0, t.Errorf("failed to read next sequence from cursor: %w", err)
	}

	return nextValue, nil
}
