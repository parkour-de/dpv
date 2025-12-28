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

	for i := 0; i < 10; i++ { // Retry a few times if collision occurs
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
