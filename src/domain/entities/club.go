package entities

import "time"

type Club struct {
	Entity
	Name                 string         `json:"name"`
	LegalForm            string         `json:"legal_form"` // e.V., GmbH, etc.
	Membership           Membership     `json:"membership"`
	Members              int            `json:"members"` // Number of members for contribution calc
	Votes                int            `json:"votes"`   // Votes in assembly, updated post-upload
	ContactPerson        string         `json:"contact_person,omitempty"`
	Email                string         `json:"email,omitempty"`
	WebsiteOK            bool           `json:"website_ok"`
	WebsiteVerification  time.Time      `json:"website_verification"`
	ParentKey            string         `json:"parent_key,omitempty"` // For recursive SubsidiaryOf edge
	OwnerKey             string         `json:"owner_key"`            // Initial creator (User key)
	StatutesOK           bool           `json:"statutes_ok,omitempty"`
	StatutesVerification time.Time      `json:"statutes_verification"`
	RegistryOK           bool           `json:"registry_ok,omitempty"`
	RegistryVerification time.Time      `json:"registry_verification"`
	Vorstand             []VorstandUser `json:"vorstand,omitempty"` // Populated via query, omitted if empty
}

func (c *Club) GetMembership() *Membership {
	return &c.Membership
}
