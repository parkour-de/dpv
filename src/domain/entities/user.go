package entities

import (
	"fmt"
	"strings"
	"time"
)

type User struct {
	Entity
	Email         string     `json:"email"`
	PasswordHash  string     `json:"password_hash"`
	LastName      string     `json:"lastname"`
	FirstName     string     `json:"firstname"`
	Roles         []string   `json:"roles"`
	EmailVerified *time.Time `json:"email_verified,omitempty"`
	Membership    Membership `json:"membership"`
	Language      string     `json:"language"`
	YourClub      string     `json:"your_club,omitempty"`
}

func (u *User) GetMembership() *Membership {
	return &u.Membership
}

func (u *User) FilteredResponse(isAdmin bool) interface{} {
	resp := &User{
		Entity: Entity{
			Key:      u.Key,
			Created:  u.Created,
			Modified: u.Modified,
		},
		Email:      u.Email,
		LastName:   u.LastName,
		FirstName:  u.FirstName,
		Roles:      u.Roles,
		Membership: u.Membership.FilteredResponse(isAdmin),
		Language:   u.Language,
		YourClub:   u.YourClub,
	}
	// We retain EmailVerified or other non-secret stuff, but intentionally drop PasswordHash
	return resp
}

func (u User) GetCSVHeaders() []string {
	return []string{"ID", "Typ", "Name", "Email", "Status", "Mitgliedsform", "Nummer", "Beitrag"}
}

func (u User) ToCSV() []string {
	name := fmt.Sprintf("%s %s", u.FirstName, u.LastName)
	memType := u.Membership.Type
	if memType == "" {
		memType = "active"
	}

	fee := fmt.Sprintf("%.2f", u.Membership.CurrentFee)
	fee = strings.ReplaceAll(fee, ".", ",")

	return []string{
		u.Key,
		"Person",
		name,
		u.Email,
		u.Membership.Status,
		memType,
		u.Membership.MembershipNumber,
		fee,
	}
}
