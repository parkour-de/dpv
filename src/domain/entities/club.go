package entities

import "time"

// Club represents a Verein or Organisation
type Club struct {
	Entity
	Name              string    `json:"name"`
	Rechtsform        string    `json:"rechtsform"`      // e.V., GmbH, etc.
	Status            string    `json:"status"`          // Aktiv, Gekündigt, etc.
	Mitgliedsstatus   string    `json:"mitgliedsstatus"` // Ordentliches Mitglied, etc.
	Mitglieder        int       `json:"mitglieder"`      // Number of members for contribution calc
	Stimmen           int       `json:"stimmen"`         // Votes in assembly, updated post-upload
	Beitrag           float64   `json:"beitrag"`         // Annual contribution
	IBAN              string    `json:"iban,omitempty"`
	SEPAMandatsnummer string    `json:"sepamandatsnummer,omitempty"`
	Ansprechpartner   string    `json:"ansprechpartner,omitempty"`
	Email             string    `json:"email,omitempty"`
	Adresse           string    `json:"adresse,omitempty"` // Combined address
	WebsiteOK         bool      `json:"website_ok"`
	WebsitePruefung   time.Time `json:"website_pruefung"`
	ParentKey         string    `json:"parent_key,omitempty"` // For recursive SubsidiaryOf edge
	OwnerKey          string    `json:"owner_key"`            // Initial creator (User key)
	// Additional flags from requirements
	SatzungOK        bool      `json:"satzung_ok,omitempty"`
	SatzungPruefung  time.Time `json:"satzung_pruefung"`
	RegisterOK       bool      `json:"register_ok,omitempty"`
	RegisterPruefung time.Time `json:"register_pruefung"`
}
