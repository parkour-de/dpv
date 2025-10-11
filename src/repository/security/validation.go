package security

import (
	"crypto"
	"crypto/hmac"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// GenerateValidationToken creates a secure token for validation actions
func GenerateValidationToken(command string, userKey string, expiry int64, parameter string, passwordHash string, secret string) (string, error) {
	data := fmt.Sprintf("%s\x01%s\x01%d\x01%s\x01%s\x01%s", command, userKey, expiry, parameter, passwordHash, secret)
	// Ensure separator appears exactly 5 times
	if strings.Count(data, "\x01") != 5 {
		return "", fmt.Errorf("separator appears %d times, expected 5", strings.Count(data, "\x01"))
	}
	hash, err := bcrypt.GenerateFromPassword(saltString(data, secret), 10)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(hash), nil
}

// ValidateToken verifies if a validation token is correct
func ValidateToken(command string, userKey string, expiry int64, parameter string, passwordHash string, secret string, token string) bool {
	data := fmt.Sprintf("%s\x01%s\x01%d\x01%s\x01%s\x01%s", command, userKey, expiry, parameter, passwordHash, secret)
	// Ensure separator appears exactly 5 times
	if strings.Count(data, "\x01") != 5 {
		return false
	}
	tokenBytes, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return false
	}
	// Check that the decoded token begins with "$2a$10$"
	if len(tokenBytes) < 7 || string(tokenBytes[:7]) != "$2a$10$" {
		return false
	}
	return bcrypt.CompareHashAndPassword(tokenBytes, saltString(data, secret)) == nil
}

func saltString(data string, secret string) []byte {
	hash := hmac.New(crypto.SHA256.New, []byte(secret))
	hash.Write([]byte(data))
	return hash.Sum(nil)
}
