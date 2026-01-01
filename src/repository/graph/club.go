package graph

import (
	"context"
	"dpv/dpv/src/domain/entities"
	"dpv/dpv/src/repository/t"

	"math"

	"github.com/arangodb/go-driver/v2/arangodb"
	"github.com/arangodb/go-driver/v2/arangodb/shared"
)

type ClubQueryOptions struct {
	Skip   int
	Limit  int
	Status string
}

// CreateClub creates a club and an 'authorizes' edge for the creator.
func (db *Db) CreateClub(ctx context.Context, club *entities.Club, userKey string) error {
	// Create the club document
	err := db.Clubs.Create(club, ctx)
	if err != nil {
		return err
	}

	// Create the 'authorizes' edge
	edge := map[string]interface{}{
		"_from": "users/" + userKey,
		"_to":   "clubs/" + club.GetKey(),
		"role":  "vorstand",
		"type":  "authorizes",
	}

	_, err = db.Edges.CreateDocument(ctx, edge)
	if err != nil {
		return t.Errorf("could not create authorization edge: %w", err)
	}

	return nil
}

// GetAdministeredClubs returns all clubs where the user is a board member.
func (db *Db) GetAdministeredClubs(ctx context.Context, userKey string) ([]entities.Club, error) {
	query := `
		FOR v, e IN 1..1 OUTBOUND @userKey edges
			FILTER e.type == "authorizes" AND e.role == "vorstand"
			RETURN v
	`
	bindVars := map[string]interface{}{
		"userKey": "users/" + userKey,
	}

	cursor, err := db.Database.Query(ctx, query, &arangodb.QueryOptions{BindVars: bindVars})
	if err != nil {
		return nil, t.Errorf("query for administered clubs failed: %w", err)
	}
	defer cursor.Close()

	var result []entities.Club
	for {
		var doc entities.Club
		_, err := cursor.ReadDocument(ctx, &doc)
		if shared.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, t.Errorf("obtaining club document failed: %w", err)
		}
		result = append(result, doc)
	}

	return result, nil
}

// GetClubByKey retrieves a club by its key including Vorstand information.
func (db *Db) GetClubByKey(ctx context.Context, key string) (*entities.Club, error) {
	query := `
		LET club = DOCUMENT("clubs", @key)
		LET vorstand = (
			FOR v, e IN 1..1 INBOUND CONCAT("clubs/", @key) edges
				FILTER e.type == "authorizes" AND e.role == "vorstand"
				RETURN {_key: v._key, firstname: v.firstname, lastname: v.lastname}
		)
		RETURN MERGE(club, {vorstand: vorstand})
	`
	bindVars := map[string]interface{}{
		"key": key,
	}

	cursor, err := db.Database.Query(ctx, query, &arangodb.QueryOptions{BindVars: bindVars})
	if err != nil {
		return nil, t.Errorf("query for club failed: %w", err)
	}
	defer cursor.Close()

	var result entities.Club
	_, err = cursor.ReadDocument(ctx, &result)
	if shared.IsNoMoreDocuments(err) {
		return nil, t.Errorf("club not found")
	} else if err != nil {
		return nil, t.Errorf("obtaining club document failed: %w", err)
	}

	result.SetKey(key)
	return &result, nil
}

// UpdateClub updates an existing club.
func (db *Db) UpdateClub(ctx context.Context, club *entities.Club) error {
	return db.Clubs.Update(club, ctx)
}

// DeleteClub deletes a club.
func (db *Db) DeleteClub(ctx context.Context, club *entities.Club) error {
	query := `
		FOR e IN edges
			FILTER e._from == @id OR e._to == @id
			REMOVE e IN edges
	`
	bindVars := map[string]interface{}{
		"id": "clubs/" + club.GetKey(),
	}

	_, err := db.Database.Query(ctx, query, &arangodb.QueryOptions{BindVars: bindVars})
	if err != nil {
		return t.Errorf("failed to remove club edges: %w", err)
	}

	return db.Clubs.Delete(club, ctx)
}

// GetClubs returns all clubs matching the options.
func (db *Db) GetClubs(ctx context.Context, options ClubQueryOptions) ([]entities.Club, error) {
	query, bindVars := buildClubQuery(options)
	cursor, err := db.Database.Query(ctx, query, &arangodb.QueryOptions{BindVars: bindVars})
	if err != nil {
		return nil, t.Errorf("query for clubs failed: %w", err)
	}
	defer cursor.Close()

	var result []entities.Club
	for {
		var doc entities.Club
		_, err := cursor.ReadDocument(ctx, &doc)
		if shared.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, t.Errorf("obtaining club document failed: %w", err)
		}
		result = append(result, doc)
	}

	return result, nil
}

func buildClubQuery(options ClubQueryOptions) (string, map[string]interface{}) {
	var query string
	bindVars := map[string]interface{}{}
	query += "FOR club IN clubs\n"
	if options.Status != "" {
		query += "  FILTER club.membership.status == @status\n"
		bindVars["status"] = options.Status
	}
	query += "  SORT club.name\n"
	if options.Skip > 0 || options.Limit > 0 {
		if options.Limit == 0 {
			options.Limit = math.MaxInt32
		}
		query += "  LIMIT @skip, @limit\n"
		bindVars["skip"] = options.Skip
		bindVars["limit"] = options.Limit
	}
	query += "  RETURN club"
	return query, bindVars
}
