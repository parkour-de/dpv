package graph

import (
	"context"
	"dpv/dpv/src/domain/entities"
	"dpv/dpv/src/repository/t"
	"strings"

	"github.com/arangodb/go-driver/v2/arangodb"
	"github.com/arangodb/go-driver/v2/arangodb/shared"
)

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

// GetUsersByFilter dynamically queries users based on membership status and existing club edges
func (db *Db) GetUsersByFilter(ctx context.Context, memStatus, hasClub string) ([]entities.User, error) {
	query := "FOR u IN users"
	var filters []string
	bindVars := map[string]interface{}{}

	if memStatus != "" {
		filters = append(filters, "u.membership.status == @memStatus")
		bindVars["memStatus"] = memStatus
	}

	switch hasClub {
	case "true":
		filters = append(filters, "LENGTH(FOR e IN edges FILTER e._from == u._id AND e.type == 'authorizes' LIMIT 1 RETURN 1) > 0")
	case "false":
		filters = append(filters, "LENGTH(FOR e IN edges FILTER e._from == u._id AND e.type == 'authorizes' LIMIT 1 RETURN 1) == 0")
	}

	if len(filters) > 0 {
		query += " FILTER " + strings.Join(filters, " AND ")
	}
	query += " SORT u.lastname, u.firstname RETURN u"

	builder := func() (string, map[string]interface{}) { return query, bindVars }
	return db.GetUsers(ctx, builder)
}
