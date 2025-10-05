package users

import (
	"context"
	"dpv/dpv/src/api"
	"dpv/dpv/src/domain/entities"
	"dpv/dpv/src/repository/security"
	"dpv/dpv/src/repository/t"
	"dpv/dpv/src/service/user"
	"encoding/json"
	"net/http"
)

type UserHandler struct {
	Service *user.Service
}

func NewHandler(service *user.Service) *UserHandler {
	return &UserHandler{Service: service}
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Vorname  string `json:"vorname"`
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.Error(w, r, t.Errorf("read request body failed: %w", err), http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" {
		api.Error(w, r, t.Errorf("email and password required"), http.StatusBadRequest)
		return
	}
	hash, err := security.HashPassword(req.Password)
	if err != nil {
		api.Error(w, r, t.Errorf("could not hash password: %w", err), http.StatusInternalServerError)
		return
	}
	userEntity := &entities.User{
		Email:        req.Email,
		PasswordHash: hash,
		Name:         req.Name,
		Vorname:      req.Vorname,
		Roles:        []string{"user"},
	}
	if err := h.Service.CreateUser(context.Background(), userEntity); err != nil {
		api.Error(w, r, t.Errorf("could not create user: %w", err), http.StatusInternalServerError)
		return
	}
	api.SuccessJson(w, r, userEntity)
}
