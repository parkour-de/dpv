package security

import (
	"crypto/rand"
	"dpv/dpv/src/repository/t"
	"encoding/base64"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 6)
	if err != nil {
		return "", err
	}
	return string(newHashedPassword), nil
}

func CheckPasswordHash(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func MakeNonce() (string, error) {
	b := make([]byte, 6)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func IsStrongPassword(password string) (bool, error) {
	var reasons []string

	if len(password) < 10 {
		reasons = append(reasons, t.T("too short (min 10 characters)"))
	}

	allDigits := true
	allLower := true
	allUpper := true
	glyphs := make(map[rune]struct{})

	for _, c := range password {
		glyphs[c] = struct{}{}
		if c < '0' || c > '9' {
			allDigits = false
		}
		if c < 'a' || c > 'z' {
			allLower = false
		}
		if c < 'A' || c > 'Z' {
			allUpper = false
		}
	}

	if allDigits {
		reasons = append(reasons, t.T("must not be only digits"))
	}
	if allLower {
		reasons = append(reasons, t.T("must not be only lowercase letters"))
	}
	if allUpper {
		reasons = append(reasons, t.T("must not be only uppercase letters"))
	}
	if len(glyphs) < 8 {
		reasons = append(reasons, t.T("must have at least 8 different glyphs"))
	}

	if len(reasons) > 0 {
		return false, errors.New(strings.Join(reasons, ", "))
	}
	return true, nil
}
