package users

import (
	"dpv/dpv/src/api"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// PaymentDetailsResponse represents payment information with role-based visibility
type PaymentDetailsResponse struct {
	IBAN              string `json:"iban"`
	AccountHolder     string `json:"account_holder,omitempty"`
	SEPAMandateNumber string `json:"sepa_mandate_number,omitempty"`
}

// GetPaymentDetails returns the current user's unmasked payment information
func (h *UserHandler) GetPaymentDetails(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	user, err := api.GetUserFromContext(r)
	if err != nil {
		api.Error(w, r, err, http.StatusUnauthorized)
		return
	}

	response := PaymentDetailsResponse{
		IBAN:              user.Membership.IBAN,
		AccountHolder:     user.Membership.AccountHolder,
		SEPAMandateNumber: user.Membership.SEPAMandateNumber,
	}

	api.SuccessJson(w, r, response)
}
