package membership

import (
	"context"
	"dpv/dpv/src/api"
	"dpv/dpv/src/repository/t"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func HandleApply(w http.ResponseWriter, r *http.Request, applyFn func(ctx context.Context, beginDate int64) error) {
	var req struct {
		BeginDate int64 `json:"begin_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		api.Error(w, r, t.Errorf("read request body failed: %w", err), http.StatusBadRequest)
		return
	}

	if err := applyFn(r.Context(), req.BeginDate); err != nil {
		api.Error(w, r, err, http.StatusBadRequest)
		return
	}

	api.SuccessJson(w, r, map[string]string{"message": t.T(t.Errorf("application submitted"), api.DetectLanguage(r))})
}

func HandleApprove(w http.ResponseWriter, r *http.Request, approveFn func(ctx context.Context, beginDate int64) error) {
	var req struct {
		BeginDate int64 `json:"begin_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		api.Error(w, r, t.Errorf("read request body failed: %w", err), http.StatusBadRequest)
		return
	}

	if err := approveFn(r.Context(), req.BeginDate); err != nil {
		api.Error(w, r, err, http.StatusBadRequest)
		return
	}

	api.SuccessJson(w, r, map[string]string{"message": t.T(t.Errorf("membership approved"), api.DetectLanguage(r))})
}

func HandleDeny(w http.ResponseWriter, r *http.Request, denyFn func(ctx context.Context) error) {
	if err := denyFn(r.Context()); err != nil {
		api.Error(w, r, err, http.StatusBadRequest)
		return
	}

	api.SuccessJson(w, r, map[string]string{"message": t.T(t.Errorf("membership denied"), api.DetectLanguage(r))})
}

func HandleCancel(w http.ResponseWriter, r *http.Request, cancelFn func(ctx context.Context, endDate int64) error) {
	var req struct {
		EndDate int64 `json:"end_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		api.Error(w, r, t.Errorf("read request body failed: %w", err), http.StatusBadRequest)
		return
	}

	if err := cancelFn(r.Context(), req.EndDate); err != nil {
		api.Error(w, r, err, http.StatusBadRequest)
		return
	}

	api.SuccessJson(w, r, map[string]string{"message": t.T(t.Errorf("membership cancelled/reset"), api.DetectLanguage(r))})
}
