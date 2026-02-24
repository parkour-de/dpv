package users

import (
	"context"
	"dpv/dpv/src/api"
	"dpv/dpv/src/endpoints/membership"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (h *UserHandler) Apply(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	user, err := api.GetUserFromContext(r)
	if err != nil {
		api.Error(w, r, err, http.StatusUnauthorized)
		return
	}

	membership.HandleApply(w, r, func(ctx context.Context, beginDate int64) error {
		return h.Service.Apply(ctx, user.Key, beginDate)
	})
}

func (h *UserHandler) Approve(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

func (h *UserHandler) Deny(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

func (h *UserHandler) Cancel(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	_, err := api.RequireGlobalAdmin(r, h.Service.DB)
	if err != nil {
		api.Error(w, r, err, http.StatusUnauthorized)
		return
	}

	key := ps.ByName("key")
	membership.HandleCancel(w, r, func(ctx context.Context, endDate int64) error {
		return h.Service.Cancel(ctx, key, endDate)
	})
}

func (h *UserHandler) CancelMe(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	user, err := api.GetUserFromContext(r)
	if err != nil {
		api.Error(w, r, err, http.StatusUnauthorized)
		return
	}

	membership.HandleCancel(w, r, func(ctx context.Context, endDate int64) error {
		return h.Service.Cancel(ctx, user.Key, endDate)
	})
}
