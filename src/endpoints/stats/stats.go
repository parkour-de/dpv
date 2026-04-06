package stats

import (
	"dpv/dpv/src/api"
	"dpv/dpv/src/service/stats"
	"net/http"
	"github.com/julienschmidt/httprouter"
)

type Handler struct {
	Service *stats.Service
}

func NewHandler(service *stats.Service) *Handler {
	return &Handler{Service: service}
}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := r.Context()
	res, err := h.Service.GetStats(ctx)
	if err != nil {
		api.Error(w, r, err, http.StatusInternalServerError)
		return
	}
	api.SuccessJson(w, r, res)
}
