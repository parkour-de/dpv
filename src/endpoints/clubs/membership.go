package clubs

import (
	"context"
	"dpv/dpv/src/api"
	"dpv/dpv/src/endpoints/membership"
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
	membership.HandleApply(w, r, func(ctx context.Context, beginDate int64, memType string, fee float64) error {
		return h.Service.Apply(ctx, key, user, beginDate, memType, fee)
	})
}

// Approve handles membership approval (Admin only).
func (h *ClubHandler) Approve(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	_, err := api.RequireGlobalAdmin(r, h.Service.DB)
	if err != nil {
		api.Error(w, r, err, http.StatusUnauthorized)
		return
	}

	key := ps.ByName("key")
	membership.HandleApprove(w, r, func(ctx context.Context, beginDate int64) error {
		return h.Service.Approve(ctx, key, beginDate)
	})
}

// Deny handles membership denial (Admin only).
func (h *ClubHandler) Deny(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	_, err := api.RequireGlobalAdmin(r, h.Service.DB)
	if err != nil {
		api.Error(w, r, err, http.StatusUnauthorized)
		return
	}

	key := ps.ByName("key")
	membership.HandleDeny(w, r, func(ctx context.Context) error {
		return h.Service.Deny(ctx, key)
	})
}

// Cancel handles membership cancellation.
func (h *ClubHandler) Cancel(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user, err := api.GetUserFromContext(r)
	if err != nil {
		api.Error(w, r, err, http.StatusUnauthorized)
		return
	}

	key := ps.ByName("key")
	membership.HandleCancel(w, r, func(ctx context.Context, endDate int64) error {
		return h.Service.Cancel(ctx, key, user, endDate)
	})
}
