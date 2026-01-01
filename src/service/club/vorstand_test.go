package club

import (
	"dpv/dpv/src/domain/entities"
	"encoding/json"
	"testing"
)

// TestVorstandJSONSerialization verifies that Vorstand field is properly serialized
func TestVorstandJSONSerialization(t *testing.T) {
	club := &entities.Club{
		Entity: entities.Entity{
			Key: "test123",
		},
		Name:      "Test Club",
		LegalForm: "e.V.",
		Vorstand: []entities.VorstandUser{
			{Key: "user1", Firstname: "John", Lastname: "Doe"},
			{Key: "user2", Firstname: "Jane", Lastname: "Smith"},
		},
	}

	// Try to marshal to JSON
	jsonData, err := json.Marshal(club)
	if err != nil {
		t.Fatalf("Failed to marshal club: %v", err)
	}

	// Unmarshal to map to check what fields are present
	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	t.Logf("JSON output: %s", string(jsonData))

	// Check if vorstand is in the output
	if _, exists := result["vorstand"]; !exists {
		t.Error("Vorstand field is missing from JSON output - json:\"-\" tag prevents serialization!")
	} else {
		t.Log("✓ Vorstand field is present in JSON output")
	}
}
