package graph

import (
	"context"
	"dpv/dpv/src/domain/entities"
	"dpv/dpv/src/repository/dpv"
	"dpv/dpv/src/repository/t"

	"github.com/arangodb/go-driver/v2/arangodb"
	"github.com/arangodb/go-driver/v2/arangodb/shared"
)

type Db struct {
	Database arangodb.Database
	Users    EntityManager[*entities.User]
	Edges    arangodb.Collection
}

func NewDB(database arangodb.Database, config *dpv.Config) (*Db, error) {
	users, err := NewEntityManager[*entities.User](database, "users", false, func() *entities.User { return new(entities.User) })
	if err != nil {
		return nil, err
	}
	edges, err := GetOrCreateCollection(database, "edges", true)
	if err != nil {
		return nil, t.Errorf("could not get or create edges collection: %w", err)
	}
	return &Db{
		database,
		users,
		edges,
	}, nil
}

// QueryBuilder defines a function that returns a query string and bind variables.
type QueryBuilder func() (string, map[string]interface{})

// buildUsersByEmailQuery returns a query and bindVars for finding users by email.
func buildUsersByEmailQuery(email string) QueryBuilder {
	query := "FOR user IN users FILTER user.email == @email RETURN user"
	bindVars := map[string]interface{}{"email": email}
	return func() (string, map[string]interface{}) { return query, bindVars }
}

// GetUsers executes a query and returns the matching users.
func (db *Db) GetUsers(ctx context.Context, builder QueryBuilder) ([]entities.User, error) {
	query, bindVars := builder()
	cursor, err := db.Database.Query(ctx, query, &arangodb.QueryOptions{BindVars: bindVars})
	if err != nil {
		return nil, t.Errorf("query string invalid: %w", err)
	}
	defer cursor.Close()

	var result []entities.User
	for {
		var doc entities.User
		_, err := cursor.ReadDocument(ctx, &doc)
		if shared.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, t.Errorf("obtaining documents failed: %w", err)
		}
		result = append(result, doc)
	}

	return result, nil
}

// GetUsersByEmail retrieves users by email using the query builder.
func (db *Db) GetUsersByEmail(ctx context.Context, email string) ([]entities.User, error) {
	return db.GetUsers(ctx, buildUsersByEmailQuery(email))
}
