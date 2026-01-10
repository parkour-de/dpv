package api

import (
	"context"
	"dpv/dpv/src/domain/entities"
	"net/http/httptest"
	"testing"
)

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		name           string
		userLanguage   string
		headerLanguage string
		acceptLanguage string
		expected       string
	}{
		{
			name:           "User preference takes precedence",
			userLanguage:   "fr",
			headerLanguage: "es",
			acceptLanguage: "de",
			expected:       "fr",
		},
		{
			name:           "X-Language header used if no user preference",
			userLanguage:   "",
			headerLanguage: "es",
			acceptLanguage: "de",
			expected:       "es",
		},
		{
			name:           "Accept-Language handled correctly (simple)",
			userLanguage:   "",
			headerLanguage: "",
			acceptLanguage: "de",
			expected:       "de",
		},
		{
			name:           "Accept-Language handled correctly (de-DE)",
			userLanguage:   "",
			headerLanguage: "",
			acceptLanguage: "de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7",
			expected:       "de",
		},
		{
			name:           "Accept-Language fallback to default",
			userLanguage:   "",
			headerLanguage: "",
			acceptLanguage: "",
			expected:       "en",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)

			if tt.userLanguage != "" {
				user := &entities.User{Language: tt.userLanguage}
				ctx := context.WithValue(req.Context(), "user", user)
				req = req.WithContext(ctx)
			}

			if tt.headerLanguage != "" {
				req.Header.Set("X-Language", tt.headerLanguage)
			}

			if tt.acceptLanguage != "" {
				req.Header.Set("Accept-Language", tt.acceptLanguage)
			}

			if got := DetectLanguage(req); got != tt.expected {
				t.Errorf("DetectLanguage() = %v, want %v", got, tt.expected)
			}
		})
	}
}
