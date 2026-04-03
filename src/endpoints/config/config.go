package config

import (
	"dpv/dpv/src/api"
	"dpv/dpv/src/service/config"
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Handler struct {
	Service *config.Service
}

func NewHandler(service *config.Service) *Handler {
	return &Handler{
		Service: service,
	}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	cfg, err := h.Service.GetConfig(r.Context())
	if err != nil {
		api.Error(w, r, err, http.StatusInternalServerError)
		return
	}
	api.SuccessJson(w, r, cfg)
}

type updateLinksRequest struct {
	Links map[string]string `json:"links"`
}

func (h *Handler) UpdateLinks(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	_, err := api.RequireGlobalAdmin(r, h.Service.DB)
	if err != nil {
		api.Error(w, r, err, http.StatusForbidden)
		return
	}

	var req updateLinksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.Error(w, r, err, http.StatusBadRequest)
		return
	}

	cfg, err := h.Service.UpdateLinks(r.Context(), req.Links)
	if err != nil {
		api.Error(w, r, err, http.StatusInternalServerError)
		return
	}
	api.SuccessJson(w, r, cfg)
}
