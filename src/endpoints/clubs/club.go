package clubs

import (
	"dpv/dpv/src/api"
	"dpv/dpv/src/domain/entities"
	"dpv/dpv/src/repository/t"
	"dpv/dpv/src/service/club"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

type ClubHandler struct {
	Service *club.Service
}

func NewHandler(service *club.Service) *ClubHandler {
	return &ClubHandler{Service: service}
}

type CreateClubRequest struct {
	Name              string `json:"name"`
	LegalForm         string `json:"legal_form"`
	Email             string `json:"email,omitempty"`
	Address           string `json:"address,omitempty"`
	State             string `json:"state,omitempty"`
	RegisterNumber    string `json:"registerNumber,omitempty"`
	ExemptionValidity string `json:"exemptionValidity,omitempty"`
}

func (h *ClubHandler) Create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	user, err := api.GetUserFromContext(r)
	if err != nil {
		api.Error(w, r, err, http.StatusUnauthorized)
		return
	}

	var req CreateClubRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.Error(w, r, t.Errorf("read request body failed: %w", err), http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.LegalForm = strings.TrimSpace(req.LegalForm)
	req.Email = strings.TrimSpace(req.Email)
	req.Address = strings.TrimSpace(req.Address)
	req.State = strings.TrimSpace(req.State)
	req.RegisterNumber = strings.TrimSpace(req.RegisterNumber)
	req.ExemptionValidity = strings.TrimSpace(req.ExemptionValidity)

	clubEntity := &entities.Club{
		Name:              req.Name,
		LegalForm:         req.LegalForm,
		State:             req.State,
		RegisterNumber:    req.RegisterNumber,
		ExemptionValidity: req.ExemptionValidity,
		Membership: entities.Membership{
			Address: req.Address,
		},
		Email: req.Email,
	}

	err = h.Service.CreateClub(r.Context(), clubEntity, user.Key)
	if err != nil {
		api.Error(w, r, t.Errorf("could not create club: %w", err), http.StatusBadRequest)
		return
	}

	api.SuccessJson(w, r, clubEntity.FilteredResponse())
}

func (h *ClubHandler) Get(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user, err := api.GetUserFromContext(r)
	if err != nil {
		api.Error(w, r, err, http.StatusUnauthorized)
		return
	}

	key := ps.ByName("key")
	club, err := h.Service.GetClub(r.Context(), key, user)
	if err != nil {
		api.Error(w, r, t.Errorf("failed to get club: %w", err), http.StatusForbidden)
		return
	}

	api.SuccessJson(w, r, club.FilteredResponse())
}

func (h *ClubHandler) Update(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user, err := api.GetUserFromContext(r)
	if err != nil {
		api.Error(w, r, err, http.StatusUnauthorized)
		return
	}

	key := ps.ByName("key")
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		api.Error(w, r, t.Errorf("invalid JSON body"), http.StatusBadRequest)
		return
	}

	for k, v := range updates {
		if s, ok := v.(string); ok {
			updates[k] = strings.TrimSpace(s)
		}
	}

	err = h.Service.UpdateClub(r.Context(), key, updates, user)
	if err != nil {
		api.Error(w, r, t.Errorf("could not update club: %w", err), http.StatusBadRequest)
		return
	}

	club, _ := h.Service.GetClub(r.Context(), key, user)
	api.SuccessJson(w, r, club.FilteredResponse())
}

func (h *ClubHandler) Delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user, err := api.GetUserFromContext(r)
	if err != nil {
		api.Error(w, r, err, http.StatusUnauthorized)
		return
	}

	key := ps.ByName("key")
	err = h.Service.DeleteClub(r.Context(), key, user)
	if err != nil {
		api.Error(w, r, t.Errorf("could not delete club: %w", err), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ClubHandler) AddOwner(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user, err := api.GetUserFromContext(r)
	if err != nil {
		api.Error(w, r, err, http.StatusUnauthorized)
		return
	}

	key := ps.ByName("key")
	var req struct {
		Email                    string `json:"email"`
		AuthorizedRepresentative bool   `json:"authorizedRepresentative"`
		Function                 string `json:"function"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.Error(w, r, t.Errorf("read request body failed: %w", err), http.StatusBadRequest)
		return
	}
	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" {
		api.Error(w, r, t.Errorf("email must not be empty"), http.StatusBadRequest)
		return
	}

	err = h.Service.AddOwner(r.Context(), key, req.Email, req.AuthorizedRepresentative, req.Function, user)
	if err != nil {
		api.Error(w, r, t.Errorf("could not add owner: %w", err), http.StatusBadRequest)
		return
	}
	// Return updated club
	club, _ := h.Service.GetClub(r.Context(), key, user)
	api.SuccessJson(w, r, club.FilteredResponse())
}

func (h *ClubHandler) RemoveOwner(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user, err := api.GetUserFromContext(r)
	if err != nil {
		api.Error(w, r, err, http.StatusUnauthorized)
		return
	}

	key := ps.ByName("key")
	targetUserKey := ps.ByName("userKey")

	err = h.Service.RemoveOwner(r.Context(), key, targetUserKey, user)
	if err != nil {
		api.Error(w, r, t.Errorf("could not remove user from list of owners: %w", err), http.StatusBadRequest)
		return
	}
	// Return updated club
	club, _ := h.Service.GetClub(r.Context(), key, user)
	api.SuccessJson(w, r, club.FilteredResponse())
}

func (h *ClubHandler) Search(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	_, err := api.GetUserFromContext(r)
	if err != nil {
		api.Error(w, r, err, http.StatusUnauthorized)
		return
	}

	query := r.URL.Query().Get("q")
	clubs, err := h.Service.SearchClubs(r.Context(), query)
	if err != nil {
		api.Error(w, r, t.Errorf("search failed: %w", err), http.StatusInternalServerError)
		return
	}

	var filtered []entities.Club
	for _, c := range clubs {
		filtered = append(filtered, *(c.FilteredResponse().(*entities.Club)))
	}

	api.SuccessJson(w, r, filtered)
}
