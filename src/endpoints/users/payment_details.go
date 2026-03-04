package users

import (
	"dpv/dpv/src/api"
	"dpv/dpv/src/repository/t"
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
		IBAN:              api.MaskIBAN(user.Membership.IBAN),
		AccountHolder:     user.Membership.AccountHolder,
		SEPAMandateNumber: "", // Non-admin gets no Mandatsreferenz
	}

	api.SuccessJson(w, r, response)
}

// GetPaymentDetailsAdmin returns the user's unmasked payment information if the requester is an admin
func (h *UserHandler) GetPaymentDetailsAdmin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	_, err := api.RequireGlobalAdmin(r, h.Service.DB)
	if err != nil {
		api.Error(w, r, err, http.StatusUnauthorized)
		return
	}

	key := ps.ByName("key")
	userEntity, err := h.Service.DB.Users.Read(key, r.Context())
	if err != nil {
		api.Error(w, r, t.Errorf("user not found"), http.StatusNotFound)
		return
	}

	response := PaymentDetailsResponse{
		IBAN:              userEntity.Membership.IBAN,
		AccountHolder:     userEntity.Membership.AccountHolder,
		SEPAMandateNumber: userEntity.Membership.SEPAMandateNumber,
	}

	api.SuccessJson(w, r, response)
}
