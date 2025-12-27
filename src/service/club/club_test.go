package club

import (
	"context"
	"dpv/dpv/src/domain/entities"
	"dpv/dpv/src/repository/graph"
	"testing"
)

func TestService_CreateClub(t *testing.T) {
	db, _, err := graph.Init("../../../config.example.yml", true)
	if err != nil {
		t.Fatalf("could not initialize database: %v", err)
	}
	defer db.Database.Remove(context.Background())

	s := NewService(db)

	club := &entities.Club{
		Name:       "Test Club",
		Rechtsform: "e.V.",
	}

	userKey := "test-user"

	err = s.CreateClub(context.Background(), club, userKey)
	if err != nil {
		t.Errorf("CreateClub failed: %v", err)
	}

	if club.GetKey() == "" {
		t.Errorf("Club key not set after creation")
	}

	// Verify authorization edge indirectly by listing administered clubs
	administered, err := s.ListClubs(context.Background(), userKey)
	if err != nil {
		t.Errorf("ListClubs failed: %v", err)
	}

	found := false
	for _, c := range administered {
		if c.GetKey() == club.GetKey() {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Club not found in administered clubs for user")
	}
}

func TestService_GetClub_Unauthorized(t *testing.T) {
	db, _, err := graph.Init("../../../config.example.yml", true)
	if err != nil {
		t.Fatalf("could not initialize database: %v", err)
	}
	defer db.Database.Remove(context.Background())

	s := NewService(db)

	club := &entities.Club{
		Name:       "Test Club",
		Rechtsform: "e.V.",
	}

	ownerKey := "owner"
	otherUser := "other"

	err = s.CreateClub(context.Background(), club, ownerKey)
	if err != nil {
		t.Fatalf("CreateClub failed: %v", err)
	}

	_, err = s.GetClub(context.Background(), club.GetKey(), otherUser)
	if err == nil {
		t.Errorf("GetClub should have failed for unauthorized user")
	}
}
