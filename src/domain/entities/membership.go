package entities

type Membership struct {
	IBAN              string  `json:"iban,omitempty"`
	SEPAMandatsnummer string  `json:"sepamandatsnummer,omitempty"`
	Beitrag           float64 `json:"beitrag"`
	Status            string  `json:"status"` // inactive, requested, active, denied, cancelled
	Adresse           string  `json:"adresse,omitempty"`
}

type MembershipProvider interface {
	GetMembership() *Membership
}
