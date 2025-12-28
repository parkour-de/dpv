package entities

type Membership struct {
	IBAN              string  `json:"iban,omitempty"`
	SEPAMandateNumber string  `json:"sepa_mandate_number,omitempty"`
	Contribution      float64 `json:"contribution"`
	Status            string  `json:"status"` // inactive, requested, active, denied, cancelled
	Address           string  `json:"address,omitempty"`
}

type MembershipProvider interface {
	GetMembership() *Membership
}
