package entities

type Membership struct {
	IBAN              string  `json:"iban,omitempty"`
	SEPAMandatsnummer string  `json:"sepamandatsnummer,omitempty"`
	Beitrag           float64 `json:"beitrag"`
	Mitgliedsstatus   string  `json:"mitgliedsstatus"` // none, requested, approved, denied, cancelled
	Adresse           string  `json:"adresse,omitempty"`
}

type MembershipProvider interface {
	GetMembership() *Membership
}
