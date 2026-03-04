package clubs

import (
	"dpv/dpv/src/api"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// PaymentDetailsResponse represents payment information with role-based visibility
type PaymentDetailsResponse struct {
	IBAN              string `json:"iban"`
	SEPAMandateNumber string `json:"sepa_mandate_number,omitempty"`
}

// GetPaymentDetails returns payment information with role-based masking
func (h *ClubHandler) GetPaymentDetails(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user, err := api.GetUserFromContext(r)
	if err != nil {
		api.Error(w, r, err, http.StatusUnauthorized)
		return
	}

	key := ps.ByName("key")

	// Fetch club using service GetClub which checks authorization
	club, err := h.Service.GetClub(r.Context(), key, user)
	if err != nil {
		api.Error(w, r, err, http.StatusForbidden)
		return
	}

	isAdmin := api.IsAdmin(*user)

	response := PaymentDetailsResponse{}

	if isAdmin {
		// Admin sees everything unmasked
		response.IBAN = club.Membership.IBAN
		response.SEPAMandateNumber = club.Membership.SEPAMandateNumber
	} else {
		// Non-admin (club owner) sees masked IBAN, no Mandatsreferenz
		response.IBAN = api.MaskIBAN(club.Membership.IBAN)
		// SEPAMandateNumber is omitted (omitempty will exclude it)
	}

	api.SuccessJson(w, r, response)
}
