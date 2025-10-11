package security

import (
	"crypto"
	"crypto/hmac"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// GenerateValidationToken creates a secure token for email validation
func GenerateValidationToken(userKey string, expiry int64, email string, passwordHash string, secret string) (string, error) {
	data := fmt.Sprintf("%s:%d:%s:%s:%s", userKey, expiry, email, passwordHash, secret)
	hash, err := bcrypt.GenerateFromPassword(saltString(data, secret), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(hash), nil
}

// ValidateToken verifies if a validation token is correct
func ValidateToken(userKey string, expiry int64, email string, passwordHash string, secret string, token string) bool {
	data := fmt.Sprintf("%s:%d:%s:%s:%s", userKey, expiry, email, passwordHash, secret)
	tokenBytes, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return false
	}
	return bcrypt.CompareHashAndPassword(tokenBytes, saltString(data, secret)) == nil
}

func saltString(data string, secret string) []byte {
	hash := hmac.New(crypto.SHA256.New, []byte(secret))
	hash.Write([]byte(data))
	return hash.Sum(nil)
}
