package security

import (
	"testing"
	"time"
)

func TestValidationToken(t *testing.T) {
	command := "validate-email"
	userKey := "user123"
	expiry := time.Now().Add(time.Hour).Unix()
	parameter := "test@example.com"
	passwordHash := "$2a$10$7R0R.Q.C.y.G.H.I.J.K.L.M.N.O.P.Q.R.S.T.U.V.W.X.Y.Z" // mock hash
	secret := "super-secret"

	token, err := GenerateValidationToken(command, userKey, expiry, parameter, passwordHash, secret)
	if err != nil {
		t.Fatalf("GenerateValidationToken failed: %v", err)
	}

	if token == "" {
		t.Fatal("token is empty")
	}

	// Test valid token
	if !ValidateToken(command, userKey, expiry, parameter, passwordHash, secret, token) {
		t.Error("ValidateToken failed for a valid token")
	}

	// Test invalid parameters
	if ValidateToken("wrong-command", userKey, expiry, parameter, passwordHash, secret, token) {
		t.Error("ValidateToken should fail for wrong command")
	}

	if ValidateToken(command, "wrong-user", expiry, parameter, passwordHash, secret, token) {
		t.Error("ValidateToken should fail for wrong user")
	}

	if ValidateToken(command, userKey, expiry+10, parameter, passwordHash, secret, token) {
		t.Error("ValidateToken should fail for wrong expiry")
	}

	if ValidateToken(command, userKey, expiry, "wrong@param.com", passwordHash, secret, token) {
		t.Error("ValidateToken should fail for wrong parameter")
	}

	if ValidateToken(command, userKey, expiry, parameter, "wrong-hash", secret, token) {
		t.Error("ValidateToken should fail for wrong password hash")
	}

	if ValidateToken(command, userKey, expiry, parameter, passwordHash, "wrong-secret", token) {
		t.Error("ValidateToken should fail for wrong secret")
	}

	if ValidateToken(command, userKey, expiry, parameter, passwordHash, secret, "invalid-token") {
		t.Error("ValidateToken should fail for invalid token string")
	}
}

func TestGenerateValidationToken_SeparatorError(t *testing.T) {
	// \x01 is the separator. Using it in parameters might cause issues if not handled.
	// Actually the current implementation just counts them.
	command := "test\x01"
	_, err := GenerateValidationToken(command, "key", 0, "param", "hash", "secret")
	if err == nil {
		t.Error("Expected error when parameter contains separator, but got nil")
	}
}
