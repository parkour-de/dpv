package dtos

import (
	"dpv/dpv/src/domain/entities"
)

// CreateClubRequest for POST /dpv/clubs
type CreateClubRequest struct {
	Name       string  `json:"name,omitempty"`
	Rechtsform string  `json:"rechtsform,omitempty"`
	Email      string  `json:"email,omitempty"`
	Adresse    string  `json:"adresse,omitempty"`
	Beitrag    float64 `json:"beitrag,omitempty"`
	ParentKey  string  `json:"parent_key,omitempty"`
}

// UpdateClubRequest for PATCH /dpv/clubs/:id (partial updates)
type UpdateClubRequest struct {
	Name       *string `json:"name,omitempty"`
	Status     *string `json:"status,omitempty"`
	Mitglieder *int    `json:"mitglieder,omitempty"`
}

// ClubResponse for GET endpoints (hides sensitive data)
type ClubResponse struct {
	entities.Entity
	Name       string  `json:"name,omitempty"`
	Rechtsform string  `json:"rechtsform,omitempty"`
	Status     string  `json:"status,omitempty"`
	Mitglieder int     `json:"mitglieder,omitempty"`
	Stimmen    int     `json:"stimmen,omitempty"`
	Beitrag    float64 `json:"beitrag,omitempty"`
	Email      string  `json:"email,omitempty"`
	WebsiteOK  bool    `json:"website_ok,omitempty"`
	ParentKey  string  `json:"parent_key,omitempty"`
}
