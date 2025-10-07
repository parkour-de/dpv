package graph

import (
	"context"
	"dpv/dpv/src/domain/entities"
	"testing"
)

func TestGetUsersByEmail(t *testing.T) {
	db, _, err := Init("../../../config.yml", true)
	if err != nil {
		t.Fatalf("db initialisation failed: %s", err)
	}

	users := []entities.User{
		{Email: "test1@example.com"},
		{Email: "test2@example.com"},
	}
	for _, user := range users {
		err := db.Users.Create(&user, context.Background())
		if err != nil {
			t.Fatalf("User creation failed: %s", err)
		}
	}

	result, err := db.GetUsersByEmail(context.Background(), "test1@example.com")
	if err != nil {
		t.Fatalf("Query failed: %s", err)
	}
	if len(result) != 1 || result[0].Email != "test1@example.com" {
		t.Errorf("Unexpected result: %+v", result)
	}
}
