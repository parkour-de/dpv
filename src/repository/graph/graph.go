package graph

import (
	"dpv/dpv/src/domain/entities"
	"dpv/dpv/src/repository/dpv"
	"dpv/dpv/src/repository/t"

	"github.com/arangodb/go-driver/v2/arangodb"
)

type Db struct {
	Database arangodb.Database
	Users    EntityManager[*entities.User]
	Clubs    EntityManager[*entities.Club]
	Edges    arangodb.Collection
	Censuses EntityManager[*entities.Census]
}

func NewDB(database arangodb.Database, config *dpv.Config) (*Db, error) {
	users, err := NewEntityManager[*entities.User](database, "users", false, func() *entities.User { return new(entities.User) })
	if err != nil {
		return nil, err
	}
	clubs, err := NewEntityManager[*entities.Club](database, "clubs", false, func() *entities.Club { return new(entities.Club) })
	if err != nil {
		return nil, err
	}
	edges, err := GetOrCreateCollection(database, "edges", true)
	if err != nil {
		return nil, t.Errorf("could not get or create edges collection: %w", err)
	}
	censuses, err := NewEntityManager[*entities.Census](database, "censuses", false, func() *entities.Census { return new(entities.Census) })
	if err != nil {
		return nil, err
	}
	return &Db{
		database,
		users,
		clubs,
		edges,
		censuses,
	}, nil
}
