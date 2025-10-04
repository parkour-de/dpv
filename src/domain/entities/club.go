package entities

import "time"

type Club struct {
	Key             string    `json:"_key,omitempty"`
	Name            string    `json:"name"`
	Rechtsform      string    `json:"rechtsform"`
	Mitgliedsstatus string    `json:"mitgliedsstatus"`
	Mitglieder      int       `json:"mitglieder"`
	Stimmen         int       `json:"stimmen"`
	SatzungOK       bool      `json:"satzung_ok"`
	SatzungPruefung time.Time `json:"satzung_pruefung"`
	// ... other fields
}
