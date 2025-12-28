package club

import (
	"context"
	"dpv/dpv/src/domain/entities"
	"dpv/dpv/src/repository/graph"
	"dpv/dpv/src/repository/storage"
	"testing"
)

func TestService_CreateClub(t *testing.T) {
	db, _, err := graph.Init("../../../config.example.yml", true)
	if err != nil {
		t.Fatalf("could not initialize database: %v", err)
	}
	defer db.Database.Remove(context.Background())

	s := NewService(db, storage.NewStorage(""))

	club := &entities.Club{
		Name:      "Test Club",
		LegalForm: "e.V.",
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

	s := NewService(db, storage.NewStorage(""))

	club := &entities.Club{
		Name:      "Test Club",
		LegalForm: "e.V.",
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

func TestService_UpdateAndDelete(t *testing.T) {
	db, _, err := graph.Init("../../../config.example.yml", true)
	if err != nil {
		t.Fatalf("could not initialize database: %v", err)
	}
	defer db.Database.Remove(context.Background())

	s := NewService(db, storage.NewStorage(""))
	ctx := context.Background()

	club := &entities.Club{
		Name:      "Initial Name",
		LegalForm: "e.V.",
	}
	userKey := "owner"

	err = s.CreateClub(ctx, club, userKey)
	if err != nil {
		t.Fatalf("CreateClub failed: %v", err)
	}
	key := club.GetKey()

	// Partial Update 1
	updates := map[string]interface{}{
		"name": "Updated Name",
	}
	err = s.UpdateClub(ctx, key, updates, userKey)
	if err != nil {
		t.Errorf("UpdateClub partial 1 failed: %v", err)
	}

	// Validate Partial Update 1
	updated, err := s.GetClub(ctx, key, userKey)
	if err != nil {
		t.Fatalf("GetClub failed after partial update 1: %v", err)
	}
	if updated.Name != "Updated Name" {
		t.Errorf("Name not updated: got %s, want %s", updated.Name, "Updated Name")
	}
	if updated.LegalForm != "e.V." {
		t.Errorf("LegalForm changed unexpectedly: got %s, want %s", updated.LegalForm, "e.V.")
	}

	// Full Update (Partial Update 2 with more fields)
	updates = map[string]interface{}{
		"name":       "Final Name",
		"legal_form": "GmbH",
		"email":      "test@example.com",
		"members":    float64(50),
		"votes":      float64(3),
	}
	err = s.UpdateClub(ctx, key, updates, userKey)
	if err != nil {
		t.Errorf("UpdateClub partial 2 failed: %v", err)
	}

	// Validate Full Update
	updated, err = s.GetClub(ctx, key, userKey)
	if err != nil {
		t.Fatalf("GetClub failed after partial update 2: %v", err)
	}
	if updated.Name != "Final Name" {
		t.Errorf("Name not updated: got %s", updated.Name)
	}
	if updated.LegalForm != "GmbH" {
		t.Errorf("LegalForm not updated: got %s", updated.LegalForm)
	}
	if updated.Email != "test@example.com" {
		t.Errorf("Email not updated: got %s", updated.Email)
	}
	// Note: Members and Votes should NOT be updated via UpdateClub
	if updated.Members != 0 {
		t.Errorf("Members SHOULD NOT be updated via PATCH: got %d, want 0", updated.Members)
	}
	if updated.Votes != 0 {
		t.Errorf("Votes SHOULD NOT be updated via PATCH: got %d, want 0", updated.Votes)
	}

	// Delete
	err = s.DeleteClub(ctx, key, userKey)
	if err != nil {
		t.Errorf("DeleteClub failed: %v", err)
	}

	// Validate Delete
	_, err = s.GetClub(ctx, key, userKey)
	if err == nil {
		t.Errorf("GetClub should have failed after deletion")
	}
}
