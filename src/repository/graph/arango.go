package graph

import (
	"context"
	"dpv/dpv/src/repository/dpv"
	"dpv/dpv/src/repository/security"
	"dpv/dpv/src/repository/t"
	"fmt"
	"log"
	"strings"

	"github.com/arangodb/go-driver/v2/arangodb"
	"github.com/arangodb/go-driver/v2/connection"
)

func Connect(config *dpv.Config, useRoot bool) (arangodb.Client, error) {
	var auth connection.Authentication
	if useRoot {
		auth = connection.NewBasicAuth("root", config.DB.Root)
	} else {
		connection.NewBasicAuth(config.DB.User, config.DB.Pass)
	}
	conn := connection.NewHttpConnection(connection.HttpConfiguration{
		Authentication: auth,
		Endpoint:       connection.NewRoundRobinEndpoints([]string{fmt.Sprintf("http://%s:%d", config.DB.Host, config.DB.Port)}),
	})

	c := arangodb.NewClient(conn)

	_, err := c.Version(context.Background())

	return c, err
}

func DropTestDatabases(c arangodb.Client) error {
	dbs, err := c.Databases(context.Background())
	if err != nil {
		return t.Errorf("could not list databases: %w", err)
	}
	for _, db := range dbs {
		if strings.HasPrefix(db.Name(), "test-") {
			err = db.Remove(context.Background())
			if err != nil {
				return t.Errorf("could not remove database: %w", err)
			}
		}
	}
	return nil
}

func GetOrCreateDatabase(c arangodb.Client, dbname string, config *dpv.Config) (arangodb.Database, error) {
	var db arangodb.Database
	if ok, err := c.DatabaseExists(context.Background(), dbname); !ok || err != nil {
		if err != nil {
			return nil, t.Errorf("failed to look for database: %w", err)
		}
		trueBool := true
		if db, err = c.CreateDatabase(context.Background(), dbname, &arangodb.CreateDatabaseOptions{Users: []arangodb.CreateDatabaseUserOptions{
			{UserName: config.DB.User, Password: config.DB.Pass, Active: &trueBool},
		}}); err != nil {
			return nil, t.Errorf("failed to create database: %w", err)
		}
	} else {
		if db, err = c.GetDatabase(context.Background(), dbname, &arangodb.GetDatabaseOptions{SkipExistCheck: false}); err != nil {
			return nil, t.Errorf("failed to open database: %w", err)
		}
	}
	return db, nil
}

func GetOrCreateCollection(db arangodb.Database, name string, edges bool) (arangodb.Collection, error) {
	if ok, err := db.CollectionExists(context.Background(), name); !ok || err != nil {
		if err != nil {
			return nil, t.Errorf("could not check if collection exists: %w", err)
		}
		if edges {
			collectionTypeEdge := arangodb.CollectionTypeEdge
			return db.CreateCollectionV2(context.Background(), name, &arangodb.CreateCollectionPropertiesV2{Type: &collectionTypeEdge})
		} else {
			collectionTypeDocument := arangodb.CollectionTypeDocument
			return db.CreateCollectionV2(context.Background(), name, &arangodb.CreateCollectionPropertiesV2{
				ComputedValues: &[]arangodb.ComputedValue{
					{
						Name:       "created",
						Expression: "RETURN DATE_ISO8601(DATE_NOW())",
						Overwrite:  true,
						ComputeOn:  []arangodb.ComputeOn{arangodb.ComputeOnInsert},
					},
					{
						Name:       "modified",
						Expression: "RETURN DATE_ISO8601(DATE_NOW())",
						Overwrite:  true,
						ComputeOn:  []arangodb.ComputeOn{arangodb.ComputeOnInsert, arangodb.ComputeOnReplace, arangodb.ComputeOnUpdate},
					},
				},
				Type: &collectionTypeDocument,
			})
		}
	} else {
		return db.GetCollection(context.Background(), name, &arangodb.GetCollectionOptions{SkipExistCheck: false})
	}
}

var fields map[string]arangodb.ArangoSearchElementProperties

func NewEntityManager[T Entity](db arangodb.Database, name string, edges bool, constructor func() T) (EntityManager[T], error) {
	collection, err := GetOrCreateCollection(db, name, edges)
	if err != nil {
		return EntityManager[T]{}, t.Errorf("could not get or create %s collection: %w", name, err)
	}
	return EntityManager[T]{collection, constructor}, nil
}

func Init(configPath string, test bool) (*Db, *dpv.Config, error) {
	var err error
	config, err := dpv.NewConfig(configPath)
	if err != nil {
		return nil, nil, t.Errorf("could not initialise config instance: %w", err)
	}
	if err := t.LoadLanguages(config); err != nil {
		log.Printf("Could not load languages: %v", err)
	}
	c, err := Connect(config, true)
	if err != nil {
		return nil, nil, t.Errorf("could not connect to database server: %w", err)
	}
	dbname := "dpv"
	if test {
		token, err := security.MakeNonce()
		if err != nil {
			return nil, nil, t.Errorf("could not create random token for test database: %w", err)
		}
		dbname = "test-" + dbname + "-" + token
		log.Printf("Using database %s\n", dbname)
	}
	database, err := GetOrCreateDatabase(c, dbname, config)
	if err != nil {
		return nil, nil, t.Errorf("could not use database: %w", err)
	}
	db, err := NewDB(database, config)
	if err != nil {
		return nil, nil, t.Errorf("could not initialise database: %w", err)
	}
	if !test {
		collection, err := db.Database.GetCollection(context.Background(), "users", &arangodb.GetCollectionOptions{SkipExistCheck: false})
		if err != nil {
			return nil, nil, t.Errorf("could not get users collection: %w", err)
		}
		count, err := collection.Count(context.Background())
		if err != nil {
			return nil, nil, t.Errorf("could not count users: %w", err)
		}
		if count == 0 {
			log.Println("Creating sample data")
			// SampleData(db)
		}
	}
	return db, config, err
}
