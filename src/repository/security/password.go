package security

import (
	"crypto/rand"
	"dpv/dpv/src/repository/t"
	"encoding/base64"

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
	if len(password) < 10 {
		return false, t.Errorf("too short (min 10 characters)")
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
		return false, t.Errorf("must not be only digits")
	}
	if allLower {
		return false, t.Errorf("must not be only lowercase letters")
	}
	if allUpper {
		return false, t.Errorf("must not be only uppercase letters")
	}
	if len(glyphs) < 8 {
		return false, t.Errorf("must have at least 8 different glyphs")
	}

	return true, nil
}
