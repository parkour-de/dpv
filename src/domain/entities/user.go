package entities

import "time"

type User struct {
	Entity
	Email         string     `json:"email"`
	PasswordHash  string     `json:"password_hash"`
	Name          string     `json:"name"`
	Vorname       string     `json:"vorname"`
	Roles         []string   `json:"roles"`
	EmailVerified *time.Time `json:"email_verified,omitempty"`
	Membership    Membership `json:"membership"`
}

func (u *User) GetMembership() *Membership {
	return &u.Membership
}
