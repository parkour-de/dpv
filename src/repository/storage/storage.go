package storage

import (
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// Storage provides methods to manage document storage.
type Storage struct {
	Root string
}

func NewStorage(root string) *Storage {
	return &Storage{Root: root}
}

// SaveDocument saves a document to a structured path and returns the unique filename.
func (s *Storage) SaveDocument(entityType, entityKey, originalFilename string, content io.Reader) (string, error) {
	entityDir := filepath.Join(s.Root, entityType, entityKey)
	if err := os.MkdirAll(entityDir, 0755); err != nil {
		return "", fmt.Errorf("could not create entity directory: %w", err)
	}

	sanitizedBase := sanitizeFilename(originalFilename)
	ext := filepath.Ext(sanitizedBase)
	baseWithoutExt := strings.TrimSuffix(sanitizedBase, ext)

	var finalPath string
	var finalFilename string

	for i := 0; i < 30; i++ { // Retry a few times if collision occurs
		randomSuffix := randomString(5)
		finalFilename = fmt.Sprintf("%s-%s%s", baseWithoutExt, randomSuffix, ext)
		finalPath = filepath.Join(entityDir, finalFilename)

		if _, err := os.Stat(finalPath); os.IsNotExist(err) {
			break
		}
	}

	f, err := os.Create(finalPath)
	if err != nil {
		return "", fmt.Errorf("could not create file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, content); err != nil {
		return "", fmt.Errorf("could not write file content: %w", err)
	}

	return finalFilename, nil
}

func sanitizeFilename(filename string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '.' || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, filename)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	rand.Read(b)
	for i := range b {
		b[i] = letters[int(b[i])%len(letters)]
	}
	return string(b)
}

// ListDocuments lists all filenames for a given entity.
func (s *Storage) ListDocuments(entityType, entityKey string) ([]string, error) {
	entityDir := filepath.Join(s.Root, entityType, entityKey)
	entries, err := os.ReadDir(entityDir)
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("could not read directory: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}

// GetDocumentPath returns the absolute path to a document, ensuring it is within the allowed directory.
func (s *Storage) GetDocumentPath(entityType, entityKey, filename string) (string, error) {
	// Sanitize filename to prevent path traversal
	cleanFilename := filepath.Base(filepath.Clean(filename))
	if cleanFilename == "." || cleanFilename == "/" {
		return "", fmt.Errorf("invalid filename")
	}

	path := filepath.Join(s.Root, entityType, entityKey, cleanFilename)

	// Check if file exists
	if _, err := os.Stat(path); err != nil {
		return "", err
	}

	return path, nil
}
