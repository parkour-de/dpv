package entities

type Membership struct {
	IBAN              string  `json:"iban,omitempty"`
	AccountHolder     string  `json:"account_holder,omitempty"`
	SEPAMandateNumber string  `json:"sepa_mandate_number,omitempty"`
	Contribution      float64 `json:"contribution"`
	Status            string  `json:"status"` // inactive, requested, active, denied, cancelled
	Address           string  `json:"address,omitempty"`
	BeginDate         int64   `json:"begin_date"` // unix seconds
	EndDate           int64   `json:"end_date"`   // unix seconds
	MembershipNumber  string  `json:"membership_number,omitempty"`
	CurrentFee        float64 `json:"current_fee,omitempty"`
	CurrentVotes      int     `json:"current_votes,omitempty"`
}

type MembershipProvider interface {
	GetMembership() *Membership
}
