package entities

import "time"

type User struct {
	Entity
	Email         string     `json:"email"`
	PasswordHash  string     `json:"password_hash"`
	LastName      string     `json:"lastname"`
	FirstName     string     `json:"firstname"`
	Roles         []string   `json:"roles"`
	EmailVerified *time.Time `json:"email_verified,omitempty"`
	Membership    Membership `json:"membership"`
}

func (u *User) GetMembership() *Membership {
	return &u.Membership
}
