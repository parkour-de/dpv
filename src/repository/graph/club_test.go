package graph

import (
	"context"
	"dpv/dpv/src/domain/entities"
	"testing"
)

func TestClubOwnerManagement(t *testing.T) {
	db, _, err := Init("../../../config.yml", true)
	if err != nil {
		t.Fatalf("db initialisation failed: %s", err)
	}
	ctx := context.Background()

	// 1. Create a User
	user := &entities.User{
		Email:     "owner_test@example.com",
		FirstName: "Owner",
		LastName:  "Test",
	}
	err = db.Users.Create(user, ctx)
	if err != nil {
		t.Fatalf("User creation failed: %s", err)
	}

	// 2. Create a Club
	club := &entities.Club{
		Name:      "Test Club Owners",
		LegalForm: "e.V.",
		Membership: entities.Membership{
			Status: "active",
		},
		OwnerKey: user.GetKey(),
	}
	// Note: CreateClub automatically adds the creator as owner
	err = db.CreateClub(ctx, club, user.GetKey())
	if err != nil {
		t.Fatalf("Club creation failed: %s", err)
	}

	// 3. Verify Initial Count (Should be 1)
	count, err := db.CountVorstand(ctx, club.GetKey())
	if err != nil {
		t.Fatalf("CountVorstand failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 owner, got %d", count)
	}

	// 4. Create another user
	user2 := &entities.User{
		Email:     "second_owner@example.com",
		FirstName: "Second",
		LastName:  "Owner",
	}
	err = db.Users.Create(user2, ctx)
	if err != nil {
		t.Fatalf("User 2 creation failed: %s", err)
	}

	// 5. Add user2 as owner
	err = db.AddVorstand(ctx, club.GetKey(), user2.GetKey())
	if err != nil {
		t.Fatalf("AddVorstand failed: %v", err)
	}

	// 6. Verify Count is 2
	count, err = db.CountVorstand(ctx, club.GetKey())
	if err != nil {
		t.Fatalf("CountVorstand failed: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 owners, got %d", count)
	}

	// 7. Verify GetClubByKey includes both
	fetchedClub, err := db.GetClubByKey(ctx, club.GetKey())
	if err != nil {
		t.Fatalf("GetClubByKey failed: %v", err)
	}
	if len(fetchedClub.Vorstand) != 2 {
		t.Errorf("Expected 2 Vorstand members in struct, got %d", len(fetchedClub.Vorstand))
	}

	// 8. Remove user2
	err = db.RemoveVorstand(ctx, club.GetKey(), user2.GetKey())
	if err != nil {
		t.Fatalf("RemoveVorstand failed: %v", err)
	}

	// 9. Verify Count is 1 again
	count, err = db.CountVorstand(ctx, club.GetKey())
	if err != nil {
		t.Fatalf("CountVorstand failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 owner, got %d", count)
	}

	// Cleanup (Optional / difficult if no delete cascades, but good practice in tests)
	// db.Clubs.Delete(club, ctx)
	// db.Users.Delete(user, ctx)
	// db.Users.Delete(user2, ctx)
	// Edges should be removed if we use DeleteClub, but users remain.
}
