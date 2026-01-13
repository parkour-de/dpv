package census

import (
	"dpv/dpv/src/api"
	"dpv/dpv/src/domain/entities"
	"dpv/dpv/src/repository/t"
	"dpv/dpv/src/service/census"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

type Handler struct {
	Service *census.Service
}

func NewHandler(service *census.Service) *Handler {
	return &Handler{
		Service: service,
	}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	clubKey := params.ByName("key")
	yearStr := params.ByName("year")

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		api.Error(w, r, t.Errorf("invalid year: %v", err), http.StatusBadRequest)
		return
	}

	user, _ := r.Context().Value("user").(*entities.User)
	result, err := h.Service.Get(r.Context(), clubKey, year, user)
	if err != nil {
		api.Error(w, r, err, http.StatusUnauthorized) // Could be 403, but Error helper handles it
		return
	}

	api.SuccessJson(w, r, result)
}

func (h *Handler) Upsert(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	clubKey := params.ByName("key")
	yearStr := params.ByName("year")

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		api.Error(w, r, t.Errorf("invalid year: %v", err), http.StatusBadRequest)
		return
	}

	// Validate current year or previous year? "reporting report year"
	// Let's assume any (valid) year is allowed for now, or maybe max current + 1.
	currentYear := time.Now().Year()
	if year < 1900 || year > currentYear+1 {
		api.Error(w, r, t.Errorf("year out of meaningful range"), http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		api.Error(w, r, t.Errorf("failed to get file: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Parse and Validate
	censusData, err := h.Service.ParseAndValidateCSV(file, year)
	if err != nil {
		api.Error(w, r, err, http.StatusBadRequest)
		return
	}

	// Persist
	user, _ := r.Context().Value("user").(*entities.User)
	err = h.Service.Upsert(r.Context(), clubKey, censusData, user)
	if err != nil {
		api.Error(w, r, err, http.StatusInternalServerError)
		return
	}

	api.SuccessJson(w, r, censusData)
}

func (h *Handler) DownloadSample(w http.ResponseWriter, r *http.Request) {
	lang := api.DetectLanguage(r)
	sampleData := h.Service.GenerateSampleCSV(lang)

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=\"census_sample.csv\"")
	w.Write(sampleData)
}
