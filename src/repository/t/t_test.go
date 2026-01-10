package t

import (
	"errors"
	"testing"
)

func TestTranslate(t *testing.T) {
	langMap := map[string]string{
		"hello %s":   "hallo %s",
		"failed: %w": "fehlgeschlagen: %w",
	}

	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "simple error",
			err:      Errorf("hello %s", "world"),
			expected: "hallo world",
		},
		{
			name:     "nested translatable error",
			err:      Errorf("failed: %w", Errorf("hello %s", "world")),
			expected: "fehlgeschlagen: hallo world",
		},
		{
			name:     "nested standard error",
			err:      Errorf("failed: %w", errors.New("something went wrong")),
			expected: "fehlgeschlagen: something went wrong",
		},
		{
			name:     "no translation found",
			err:      Errorf("unknown %s", "thing"),
			expected: "unknown thing",
		},
		{
			name:     "nil error",
			err:      nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Translate(tt.err, langMap)
			if got != tt.expected {
				t.Errorf("Translate() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestT(t *testing.T) {
	languages["de"] = map[string]string{"hello %s": "hallo %s"}
	got := T(Errorf("hello %s", "world"), "de")
	if got != "hallo world" {
		t.Errorf("T() = %v, want %v", got, "hallo world")
	}
}

func TestGetMapFor(t *testing.T) {
	languages["de"] = map[string]string{"key": "wert"}
	languages["fr"] = map[string]string{"key": "valeur"}

	tests := []struct {
		lang     string
		expected string
	}{
		{"de", "wert"},
		{"de-DE", "wert"},
		{"fr", "valeur"},
		{"en", ""},
	}

	for _, tt := range tests {
		m := GetMapFor(tt.lang)
		if tt.expected == "" {
			if len(m) != 0 {
				t.Errorf("GetMapFor(%s) should return empty map", tt.lang)
			}
		} else {
			if m["key"] != tt.expected {
				t.Errorf("GetMapFor(%s)[key] = %v, want %v", tt.lang, m["key"], tt.expected)
			}
		}
	}
}
