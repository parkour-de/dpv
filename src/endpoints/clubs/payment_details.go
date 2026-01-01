package clubs

import (
	"dpv/dpv/src/api"
	"net/http"
	"strings"

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
		response.IBAN = maskIBAN(club.Membership.IBAN)
		// SEPAMandateNumber is omitted (omitempty will exclude it)
	}

	api.SuccessJson(w, r, response)
}

// maskIBAN masks an IBAN showing only first 4 and last 3 alphanumeric characters
func maskIBAN(iban string) string {
	if iban == "" {
		return ""
	}

	// Remove spaces for processing
	cleaned := strings.ReplaceAll(iban, " ", "")

	if len(cleaned) <= 7 {
		// Too short to mask meaningfully, just return masked version
		return strings.Repeat("*", len(cleaned))
	}

	// Get first 4 and last 3
	first4 := cleaned[:4]
	last3 := cleaned[len(cleaned)-3:]

	// Calculate middle length
	middleLen := len(cleaned) - 7

	// Create masked version with spaces for readability (every 4 chars)
	masked := first4 + " "
	remaining := middleLen
	for remaining > 0 {
		if remaining >= 4 {
			masked += "**** "
			remaining -= 4
		} else {
			masked += strings.Repeat("*", remaining) + " "
			remaining = 0
		}
	}
	masked += last3

	return strings.TrimSpace(masked)
}
