package users

import (
	"context"
	"dpv/dpv/src/api"
	"dpv/dpv/src/domain/entities"
	"dpv/dpv/src/repository/t"
	"dpv/dpv/src/service/user"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
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

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.Error(w, r, t.Errorf("read request body failed: %w", err), http.StatusBadRequest)
		return
	}

	userEntity := &entities.User{
		Email:   req.Email,
		Name:    req.Name,
		Vorname: req.Vorname,
		Roles:   []string{"user"},
	}

	err := h.Service.CreateUser(context.Background(), userEntity, req.Password)
	if err != nil {
		// Map validation errors to 400, others to 500
		switch err.Error() {
		case t.T("vorname must not be empty"), t.T("name must not be empty"), t.T("email must not be empty"), t.T("password must not be empty"):
			api.Error(w, r, err, http.StatusBadRequest)
			return
		}
		if err.Error() == t.T("user with this email already exists") ||
			strings.Contains(err.Error(), t.T("too short (min 10 characters)")) ||
			strings.Contains(err.Error(), t.T("must not be only digits")) ||
			strings.Contains(err.Error(), t.T("must not be only lowercase letters")) ||
			strings.Contains(err.Error(), t.T("must not be only uppercase letters")) ||
			strings.Contains(err.Error(), t.T("must have at least 8 different glyphs")) {
			api.Error(w, r, err, http.StatusBadRequest)
			return
		}
		api.Error(w, r, t.Errorf("could not create user: %w", err), http.StatusInternalServerError)
		return
	}
	api.SuccessJson(w, r, userEntity)
}

// Me returns the current user
func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	userEntity, ok := r.Context().Value("user").(*entities.User)
	if !ok || userEntity == nil {
		api.Error(w, r, t.Errorf("user not found in context"), http.StatusUnauthorized)
		return
	}
	// Copy userEntity without password hash
	resp := &entities.User{
		Key:     userEntity.Key,
		Email:   userEntity.Email,
		Name:    userEntity.Name,
		Vorname: userEntity.Vorname,
		Roles:   userEntity.Roles,
	}
	api.SuccessJson(w, r, resp)
}
