package storage

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello.txt", "hello.txt"},
		{"Hello World.txt", "Hello_World.txt"},
		{"../secrets.txt", ".._secrets.txt"},
		{"file%with$symbols.pdf", "file_with_symbols.pdf"},
		{"äöü.jpg", "äöü.jpg"},
	}

	for _, tt := range tests {
		result := sanitizeFilename(tt.input)
		if result != tt.expected {
			t.Errorf("sanitizeFilename(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestRandomString(t *testing.T) {
	s1 := randomString(10)
	s2 := randomString(10)

	if len(s1) != 10 {
		t.Errorf("Expected length 10, got %d", len(s1))
	}

	if s1 == s2 {
		t.Error("randomString generated identical strings")
	}
}

func TestSaveDocument(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "storage-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	s := NewStorage(tempDir)
	content := []byte("hello world")
	entityType := "clubs"
	entityKey := "club1"
	filename := "test.txt"

	finalFilename, err := s.SaveDocument(entityType, entityKey, filename, bytes.NewReader(content))
	if err != nil {
		t.Fatalf("SaveDocument failed: %v", err)
	}

	if !strings.HasPrefix(finalFilename, "test-") || !strings.HasSuffix(finalFilename, ".txt") {
		t.Errorf("Unexpected final filename format: %s", finalFilename)
	}

	// Verify file existence and content
	expectedPath := filepath.Join(tempDir, entityType, entityKey, finalFilename)
	data, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	if string(data) != string(content) {
		t.Errorf("Expected content %q, got %q", string(content), string(data))
	}
}

func TestSaveDocument_MkdirError(t *testing.T) {
	// Use a path that cannot be created (e.g., a file where a directory should be)
	tempFile, err := os.CreateTemp("", "mkdir-collision")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	s := NewStorage(tempFile.Name())
	_, err = s.SaveDocument("type", "key", "file.txt", strings.NewReader("content"))
	if err == nil {
		t.Error("Expected error when MkdirAll fails, but got nil")
	}
}
