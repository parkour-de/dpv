package clubs

import (
	"dpv/dpv/src/api"
	"dpv/dpv/src/domain/entities"
	"dpv/dpv/src/repository/t"
	"dpv/dpv/src/service/club"
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type ClubHandler struct {
	Service *club.Service
}

func NewHandler(service *club.Service) *ClubHandler {
	return &ClubHandler{Service: service}
}

type CreateClubRequest struct {
	Name            string `json:"name"`
	Rechtsform      string `json:"rechtsform"`
	Mitgliedsstatus string `json:"mitgliedsstatus,omitempty"`
	Email           string `json:"email,omitempty"`
	Adresse         string `json:"adresse,omitempty"`
}

func (h *ClubHandler) Create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	user, ok := r.Context().Value("user").(*entities.User)
	if !ok || user == nil {
		api.Error(w, r, t.Errorf("user not found in context"), http.StatusUnauthorized)
		return
	}

	var req CreateClubRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.Error(w, r, t.Errorf("read request body failed: %w", err), http.StatusBadRequest)
		return
	}

	clubEntity := &entities.Club{
		Name:            req.Name,
		Rechtsform:      req.Rechtsform,
		Mitgliedsstatus: req.Mitgliedsstatus,
		Email:           req.Email,
		Adresse:         req.Adresse,
	}

	err := h.Service.CreateClub(r.Context(), clubEntity, user.Key)
	if err != nil {
		api.Error(w, r, err, http.StatusBadRequest)
		return
	}

	api.SuccessJson(w, r, filteredResponse(clubEntity))
}

func (h *ClubHandler) List(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	user, ok := r.Context().Value("user").(*entities.User)
	if !ok || user == nil {
		api.Error(w, r, t.Errorf("user not found in context"), http.StatusUnauthorized)
		return
	}

	clubs, err := h.Service.ListClubs(r.Context(), user.Key)
	if err != nil {
		api.Error(w, r, err, http.StatusInternalServerError)
		return
	}

	var resp []entities.Club
	for _, c := range clubs {
		resp = append(resp, *filteredResponse(&c))
	}

	api.SuccessJson(w, r, resp)
}

func (h *ClubHandler) Get(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user, ok := r.Context().Value("user").(*entities.User)
	if !ok || user == nil {
		api.Error(w, r, t.Errorf("user not found in context"), http.StatusUnauthorized)
		return
	}

	key := ps.ByName("key")
	club, err := h.Service.GetClub(r.Context(), key, user.Key)
	if err != nil {
		api.Error(w, r, err, http.StatusForbidden)
		return
	}

	api.SuccessJson(w, r, filteredResponse(club))
}

func (h *ClubHandler) Update(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user, ok := r.Context().Value("user").(*entities.User)
	if !ok || user == nil {
		api.Error(w, r, t.Errorf("user not found in context"), http.StatusUnauthorized)
		return
	}

	key := ps.ByName("key")
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		api.Error(w, r, t.Errorf("invalid JSON body"), http.StatusBadRequest)
		return
	}

	err := h.Service.UpdateClub(r.Context(), key, updates, user.Key)
	if err != nil {
		api.Error(w, r, err, http.StatusBadRequest)
		return
	}

	club, _ := h.Service.GetClub(r.Context(), key, user.Key)
	api.SuccessJson(w, r, filteredResponse(club))
}

func (h *ClubHandler) Delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user, ok := r.Context().Value("user").(*entities.User)
	if !ok || user == nil {
		api.Error(w, r, t.Errorf("user not found in context"), http.StatusUnauthorized)
		return
	}

	key := ps.ByName("key")
	err := h.Service.DeleteClub(r.Context(), key, user.Key)
	if err != nil {
		api.Error(w, r, err, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func filteredResponse(clubEntity *entities.Club) *entities.Club {
	// For Phase 3.1, suppress sensitive fields like IBAN if desired,
	// but here we return most fields except internal ones if needed.
	resp := &entities.Club{
		Entity: entities.Entity{
			Key:      clubEntity.Key,
			Created:  clubEntity.Created,
			Modified: clubEntity.Modified,
		},
		Name:             clubEntity.Name,
		Rechtsform:       clubEntity.Rechtsform,
		Status:           clubEntity.Status,
		Mitgliedsstatus:  clubEntity.Mitgliedsstatus,
		Mitglieder:       clubEntity.Mitglieder,
		Stimmen:          clubEntity.Stimmen,
		Beitrag:          clubEntity.Beitrag,
		Ansprechpartner:  clubEntity.Ansprechpartner,
		Email:            clubEntity.Email,
		Adresse:          clubEntity.Adresse,
		WebsiteOK:        clubEntity.WebsiteOK,
		WebsitePruefung:  clubEntity.WebsitePruefung,
		ParentKey:        clubEntity.ParentKey,
		OwnerKey:         clubEntity.OwnerKey,
		SatzungOK:        clubEntity.SatzungOK,
		SatzungPruefung:  clubEntity.SatzungPruefung,
		RegisterOK:       clubEntity.RegisterOK,
		RegisterPruefung: clubEntity.RegisterPruefung,
	}
	// Note: IBAN and SEPAMandatsnummer are omitted here for security/privacy in general views
	return resp
}
