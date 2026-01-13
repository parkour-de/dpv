package graph

import (
	"context"
	"dpv/dpv/src/domain/entities"
	"testing"
)

func TestCensusRepository(t_test *testing.T) {
	db, _, err := Init("../../../config.yml", true)
	if err != nil {
		t_test.Fatalf("db initialisation failed: %s", err)
	}
	ctx := context.Background()

	// 1. Create a Club
	club := &entities.Club{
		Name:      "Test Club Census",
		LegalForm: "e.V.",
	}
	err = db.Clubs.Create(club, ctx)
	if err != nil {
		t_test.Fatalf("Club creation failed: %s", err)
	}

	// 2. Upsert a Census
	census := &entities.Census{
		Year: 2024,
		Members: []entities.MemberRow{
			{Firstname: "Erika", Lastname: "Mustermann", BirthYear: 1990, Gender: "w"},
			{Firstname: "Max", Lastname: "Mustermann", BirthYear: 1985, Gender: "m"},
		},
	}
	err = db.UpsertCensus(ctx, club.GetKey(), census)
	if err != nil {
		t_test.Fatalf("UpsertCensus failed: %s", err)
	}

	// 3. Verify Census Node was created and has MemberCount
	fetchedCensus, err := db.GetCensus(ctx, club.GetKey(), 2024)
	if err != nil {
		t_test.Fatalf("GetCensus failed: %s", err)
	}
	if fetchedCensus.MemberCount != 2 {
		t_test.Errorf("expected member count 2, got %d", fetchedCensus.MemberCount)
	}

	// 4. Verify Club Summary
	fetchedClub, err := db.GetClubByKey(ctx, club.GetKey())
	if err != nil {
		t_test.Fatalf("GetClubByKey failed: %v", err)
	}
	if len(fetchedClub.Census) != 1 {
		t_test.Errorf("expected 1 census summary, got %d", len(fetchedClub.Census))
	} else {
		if fetchedClub.Census[0].Year != 2024 {
			t_test.Errorf("expected year 2024, got %d", fetchedClub.Census[0].Year)
		}
		if fetchedClub.Census[0].Count != 2 {
			t_test.Errorf("expected count 2, got %d", fetchedClub.Census[0].Count)
		}
	}

	// 5. Update Census (Add a member)
	census.Members = append(census.Members, entities.MemberRow{Firstname: "John", Lastname: "Doe", BirthYear: 2000, Gender: "m"})
	err = db.UpsertCensus(ctx, club.GetKey(), census)
	if err != nil {
		t_test.Fatalf("UpsertCensus (update) failed: %s", err)
	}

	// 6. Verify updated count
	fetchedCensus, _ = db.GetCensus(ctx, club.GetKey(), 2024)
	if fetchedCensus.MemberCount != 3 {
		t_test.Errorf("expected member count 3, got %d", fetchedCensus.MemberCount)
	}

	fetchedClub, _ = db.GetClubByKey(ctx, club.GetKey())
	if fetchedClub.Census[0].Count != 3 {
		t_test.Errorf("expected updated club summary count 3, got %d", fetchedClub.Census[0].Count)
	}
}
