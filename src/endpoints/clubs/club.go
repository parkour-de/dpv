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
	Name      string `json:"name"`
	LegalForm string `json:"legal_form"`
	Email     string `json:"email,omitempty"`
	Address   string `json:"address,omitempty"`
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

	clubEntity := &entities.Club{
		Name:      req.Name,
		LegalForm: req.LegalForm,
		Membership: entities.Membership{
			Address: req.Address,
		},
		Email: req.Email,
	}

	err = h.Service.CreateClub(r.Context(), clubEntity, user.Key)
	if err != nil {
		api.Error(w, r, err, http.StatusBadRequest)
		return
	}

	api.SuccessJson(w, r, FilteredResponse(clubEntity))
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
		api.Error(w, r, err, http.StatusForbidden)
		return
	}

	api.SuccessJson(w, r, FilteredResponse(club))
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
		api.Error(w, r, err, http.StatusBadRequest)
		return
	}

	club, _ := h.Service.GetClub(r.Context(), key, user)
	api.SuccessJson(w, r, FilteredResponse(club))
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
		api.Error(w, r, err, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func FilteredResponse(clubEntity *entities.Club) *entities.Club {
	// For Phase 3.1, suppress sensitive fields like IBAN if desired,
	// but here we return most fields except internal ones if needed.
	resp := &entities.Club{
		Entity: entities.Entity{
			Key:      clubEntity.Key,
			Created:  clubEntity.Created,
			Modified: clubEntity.Modified,
		},
		Name:      clubEntity.Name,
		LegalForm: clubEntity.LegalForm,
		Membership: entities.Membership{
			Status:       clubEntity.Membership.Status,
			Contribution: clubEntity.Membership.Contribution,
			Address:      clubEntity.Membership.Address,
		},
		Members:              clubEntity.Members,
		Votes:                clubEntity.Votes,
		ContactPerson:        clubEntity.ContactPerson,
		Email:                clubEntity.Email,
		WebsiteOK:            clubEntity.WebsiteOK,
		WebsiteVerification:  clubEntity.WebsiteVerification,
		ParentKey:            clubEntity.ParentKey,
		OwnerKey:             clubEntity.OwnerKey,
		StatutesOK:           clubEntity.StatutesOK,
		StatutesVerification: clubEntity.StatutesVerification,
		RegistryOK:           clubEntity.RegistryOK,
		RegistryVerification: clubEntity.RegistryVerification,
		Vorstand:             clubEntity.Vorstand, // Include Vorstand info from query
	}
	// Note: IBAN and SEPAMandatsnummer are omitted here for security/privacy in general views
	return resp
}
