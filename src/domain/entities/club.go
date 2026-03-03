package entities

import (
	"fmt"
	"time"
)

type Club struct {
	Entity
	Name                 string          `json:"name"`
	LegalForm            string          `json:"legal_form"` // e.V., GmbH, etc.
	Membership           Membership      `json:"membership"`
	Members              int             `json:"members"` // Number of members for contribution calc
	Votes                int             `json:"votes"`   // Votes in assembly, updated post-upload
	ContactPerson        string          `json:"contact_person,omitempty"`
	Email                string          `json:"email,omitempty"`
	State                string          `json:"state,omitempty"`
	RegisterNumber       string          `json:"registerNumber,omitempty"`
	ExemptionValidity    string          `json:"exemptionValidity,omitempty"`
	WebsiteOK            bool            `json:"website_ok"`
	WebsiteVerification  time.Time       `json:"website_verification"`
	ParentKey            string          `json:"parent_key,omitempty"` // For recursive SubsidiaryOf edge
	OwnerKey             string          `json:"owner_key"`            // Initial creator (User key)
	StatutesOK           bool            `json:"statutes_ok,omitempty"`
	StatutesVerification time.Time       `json:"statutes_verification"`
	RegistryOK           bool            `json:"registry_ok,omitempty"`
	RegistryVerification time.Time       `json:"registry_verification"`
	Vorstand             []VorstandUser  `json:"vorstand,omitempty"` // Populated via query, omitted if empty
	Census               []CensusSummary `json:"census,omitempty"`   // Populated via query, omitted if empty
}

type CensusSummary struct {
	Year  int `json:"year"`
	Count int `json:"count"`
}

func (c *Club) GetMembership() *Membership {
	return &c.Membership
}

func (c Club) GetCSVHeaders() []string {
	return []string{"ID", "Typ", "Name", "Email", "Status", "Mitgliedsform", "Nummer", "Beitrag"}
}

func (c Club) ToCSV() []string {
	memType := c.Membership.Type
	if memType == "" {
		memType = "ordinary"
	}

	fee := fmt.Sprintf("%.2f", c.Membership.CurrentFee)
	email := c.Email
	if email == "" {
		email = "No Email"
	}

	return []string{
		c.Key,
		"Club",
		c.Name,
		email,
		c.Membership.Status,
		memType,
		c.Membership.MembershipNumber,
		fee,
	}
}
