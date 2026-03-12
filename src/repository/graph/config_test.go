package graph_test

import (
	"context"
	"dpv/dpv/src/repository/graph"
	"path/filepath"
	"testing"
)

func TestGetNextSequence(t *testing.T) {
	// Setup DB connection
	configPath := filepath.Join("../../../", "cfg", "config.yml")
	db, config, err := graph.Init(configPath, true)
	if err != nil {
		t.Fatalf("could not initialise database: %v", err)
	}

	c, err := graph.Connect(config, true)
	if err != nil {
		t.Fatalf("could not connect for cleanup: %v", err)
	}

	ctx := context.Background()

	// Initial fetch
	seq1, err := db.GetNextSequence(ctx, "test_seq")
	if err != nil {
		t.Fatalf("failed to get first sequence: %v", err)
	}

	if seq1 != 1 {
		t.Errorf("Expected first sequence to be 1, got %d", seq1)
	}

	// Second fetch
	seq2, err := db.GetNextSequence(ctx, "test_seq")
	if err != nil {
		t.Fatalf("failed to get second sequence: %v", err)
	}

	if seq2 != 2 {
		t.Errorf("Expected second sequence to be 2, got %d", seq2)
	}

	// Fetch another sequence
	seq3, err := db.GetNextSequence(ctx, "other_seq")
	if err != nil {
		t.Fatalf("failed to get other sequence: %v", err)
	}

	if seq3 != 1 {
		t.Errorf("Expected other sequence to be 1, got %d", seq3)
	}

	// Cleanup
	err = graph.DropTestDatabases(c)
	if err != nil {
		t.Logf("Cleanup failed: %v", err)
	}
}
