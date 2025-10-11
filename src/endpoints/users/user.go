package users

import (
	"context"
	"dpv/dpv/src/api"
	"dpv/dpv/src/domain/entities"
	"dpv/dpv/src/repository/t"
	"dpv/dpv/src/service/user"
	"encoding/json"
	"net/http"
	"strconv"
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

// RequestEmailValidation - requires authentication
func (h *UserHandler) RequestEmailValidation(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	userEntity, ok := r.Context().Value("user").(*entities.User)
	if !ok || userEntity == nil {
		api.Error(w, r, t.Errorf("user not found in context"), http.StatusUnauthorized)
		return
	}

	var req struct {
		Email string `json:"email,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.Error(w, r, t.Errorf("read request body failed: %w", err), http.StatusBadRequest)
		return
	}

	err := h.Service.RequestEmailValidation(r.Context(), req.Email)
	if err != nil {
		api.Error(w, r, err, http.StatusBadRequest)
		return
	}

	targetEmail := req.Email
	if targetEmail == "" {
		targetEmail = userEntity.Email
	}

	api.SuccessJson(w, r, map[string]string{
		"message": t.Sprintf("Validation email sent to %s", targetEmail),
	})
}

// ValidateEmail - public endpoint, no authentication required
func (h *UserHandler) ValidateEmail(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	userKey := r.URL.Query().Get("key")
	expiryStr := r.URL.Query().Get("expiry")
	email := r.URL.Query().Get("email")
	token := r.URL.Query().Get("token")

	if userKey == "" || expiryStr == "" || email == "" || token == "" {
		api.Error(w, r, t.Errorf("missing required parameters"), http.StatusBadRequest)
		return
	}

	expiry, err := strconv.ParseInt(expiryStr, 10, 64)
	if err != nil {
		api.Error(w, r, t.Errorf("invalid expiry timestamp"), http.StatusBadRequest)
		return
	}

	err = h.Service.ValidateEmail(context.Background(), userKey, expiry, email, token)
	if err != nil {
		api.Error(w, r, err, http.StatusBadRequest)
		return
	}

	html := `<!DOCTYPE html>
<html lang="de">
<head>
    <meta charset="UTF-8">
    <title>E-Mail bestätigt - DPV</title>
    <style>body { font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; text-align: center; }</style>
</head>
<body>
    <h1>✅ E-Mail-Adresse erfolgreich bestätigt!</h1>
    <p>Ihre E-Mail-Adresse wurde erfolgreich bestätigt. Sie können jetzt alle Funktionen der DPV-Mitgliederverwaltung nutzen.</p>
    <p><a href="/">Zurück zur Startseite</a></p>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	api.Success(w, r, []byte(html))
}
