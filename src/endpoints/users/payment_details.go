package users

import (
	"dpv/dpv/src/api"
	"dpv/dpv/src/domain/entities"
	"dpv/dpv/src/repository/t"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// PaymentDetailsResponse represents payment information with role-based visibility
type PaymentDetailsResponse struct {
	IBAN              string `json:"iban"`
	AccountHolder     string `json:"account_holder"`
	SEPAMandateNumber string `json:"sepa_mandate_number"`
}

// GetPaymentDetails returns payment information with role-based masking
func (h *UserHandler) GetPaymentDetails(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user, err := api.GetUserFromContext(r)
	if err != nil {
		api.Error(w, r, err, http.StatusUnauthorized)
		return
	}

	key := ps.ByName("key")
	var targetUser *entities.User
	isAdmin := api.IsAktivAdmin(*user)

	if key == "" || key == "me" || key == user.Key {
		targetUser = user
	} else {
		if !isAdmin {
			api.Error(w, r, fmt.Errorf("forbidden"), http.StatusForbidden)
			return
		}
		targetUser, err = h.Service.DB.Users.Read(key, r.Context())
		if err != nil {
			api.Error(w, r, t.Errorf("user not found"), http.StatusNotFound)
			return
		}
	}

	response := PaymentDetailsResponse{
		AccountHolder:     targetUser.Membership.AccountHolder,
		SEPAMandateNumber: targetUser.Membership.SEPAMandateNumber,
	}

	if isAdmin {
		// Admin sees everything unmasked
		response.IBAN = targetUser.Membership.IBAN
	} else {
		// Non-admin sees masked IBAN
		response.IBAN = api.MaskIBAN(targetUser.Membership.IBAN)
	}

	api.SuccessJson(w, r, response)
}
