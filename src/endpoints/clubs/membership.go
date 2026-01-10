package clubs

import (
	"dpv/dpv/src/api"
	"dpv/dpv/src/repository/t"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Apply handles membership application.
func (h *ClubHandler) Apply(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user, err := api.GetUserFromContext(r)
	if err != nil {
		api.Error(w, r, err, http.StatusUnauthorized)
		return
	}

	key := ps.ByName("key")
	err = h.Service.Apply(r.Context(), key, user)
	if err != nil {
		api.Error(w, r, err, http.StatusBadRequest)
		return
	}

	api.SuccessJson(w, r, map[string]string{"message": t.T(t.Errorf("application submitted"), api.DetectLanguage(r))})
}

// Approve handles membership approval (Admin only).
func (h *ClubHandler) Approve(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	_, err := api.RequireGlobalAdmin(r, h.Service.DB)
	if err != nil {
		api.Error(w, r, err, http.StatusUnauthorized)
		return
	}

	key := ps.ByName("key")
	err = h.Service.Approve(r.Context(), key)
	if err != nil {
		api.Error(w, r, err, http.StatusBadRequest)
		return
	}

	api.SuccessJson(w, r, map[string]string{"message": t.T(t.Errorf("membership approved"), api.DetectLanguage(r))})
}

// Deny handles membership denial (Admin only).
func (h *ClubHandler) Deny(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	_, err := api.RequireGlobalAdmin(r, h.Service.DB)
	if err != nil {
		api.Error(w, r, err, http.StatusUnauthorized)
		return
	}

	key := ps.ByName("key")
	err = h.Service.Deny(r.Context(), key)
	if err != nil {
		api.Error(w, r, err, http.StatusBadRequest)
		return
	}

	api.SuccessJson(w, r, map[string]string{"message": t.T(t.Errorf("membership denied"), api.DetectLanguage(r))})
}

// Cancel handles membership cancellation.
func (h *ClubHandler) Cancel(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user, err := api.GetUserFromContext(r)
	if err != nil {
		api.Error(w, r, err, http.StatusUnauthorized)
		return
	}

	key := ps.ByName("key")
	err = h.Service.Cancel(r.Context(), key, user)
	if err != nil {
		api.Error(w, r, err, http.StatusBadRequest)
		return
	}

	api.SuccessJson(w, r, map[string]string{"message": t.T(t.Errorf("membership cancelled/reset"), api.DetectLanguage(r))})
}
