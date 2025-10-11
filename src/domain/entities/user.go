package entities

import "time"

type User struct {
	Entity
	Key           string     `json:"_key,omitempty"`
	Email         string     `json:"email"`
	PasswordHash  string     `json:"password_hash"`
	Name          string     `json:"name"`
	Vorname       string     `json:"vorname"`
	Roles         []string   `json:"roles"`
	EmailVerified *time.Time `json:"email_verified,omitempty"`
	// Address fields...
}
