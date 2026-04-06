package entities

type Membership struct {
	IBAN              string  `json:"iban,omitempty"`
	AccountHolder     string  `json:"account_holder,omitempty"`
	SEPAMandateNumber string  `json:"sepa_mandate_number,omitempty"`
	Type              string  `json:"type,omitempty"` // active, supporting, ordinary
	Contribution      float64 `json:"contribution"`
	Status            string  `json:"status"` // inactive, requested, active, denied, cancelled
	Address           string  `json:"address,omitempty"`
	ApplicationDate   int64   `json:"application_date,omitempty"` // unix seconds
	BeginDate         int64   `json:"begin_date"`                 // unix seconds
	EndDate           int64   `json:"end_date"`                   // unix seconds
	MembershipNumber  string  `json:"membership_number,omitempty"`
	CurrentFee        float64 `json:"current_fee,omitempty"`
	CurrentVotes      int     `json:"current_votes,omitempty"`
}

type MembershipProvider interface {
	GetMembership() *Membership
}

type FilteredResponder interface {
	FilteredResponse() interface{}
}

func (m Membership) FilteredResponse() Membership {
	resp := Membership{
		Status:           m.Status,
		Contribution:     m.Contribution,
		Type:             m.Type,
		BeginDate:        m.BeginDate,
		ApplicationDate:  m.ApplicationDate,
		EndDate:          m.EndDate,
		Address:          m.Address,
		MembershipNumber: m.MembershipNumber,
		CurrentFee:       m.CurrentFee,
		CurrentVotes:     m.CurrentVotes,
	}
	return resp
}
