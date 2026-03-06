package users

import (
	"context"
	"dpv/dpv/src/api"
	"dpv/dpv/src/domain/entities"
	"dpv/dpv/src/repository/t"
	"dpv/dpv/src/service/user"
	"encoding/json"
	"errors"
	"fmt"
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
	Email          string `json:"email"`
	Password       string `json:"password"`
	LastName       string `json:"lastname"`
	FirstName      string `json:"firstname"`
	DateOfBirth    string `json:"dateOfBirth"`
	ConsentPrivacy bool   `json:"consent_privacy"`
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.Error(w, r, t.Errorf("read request body failed: %w", err), http.StatusBadRequest)
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)

	if !req.ConsentPrivacy {
		api.Error(w, r, t.Errorf("you must accept the privacy policy"), http.StatusBadRequest)
		return
	}

	userEntity := &entities.User{
		Email:       req.Email,
		LastName:    req.LastName,
		FirstName:   req.FirstName,
		DateOfBirth: req.DateOfBirth,
		Roles:       []string{"user"},
	}

	err := h.Service.CreateUser(context.Background(), userEntity, req.Password)
	if err != nil {
		// Map validation errors to 400, others to 500
		var tErr *t.TranslatableError
		if errors.As(err, &tErr) {
			switch tErr.Key {
			case "firstname must not be empty", "lastname must not be empty", "email must not be empty", "password must not be empty", "user with this email already exists", "you must accept the privacy policy":
				api.Error(w, r, err, http.StatusBadRequest)
				return
			}
			if strings.Contains(tErr.Key, "too short") ||
				strings.Contains(tErr.Key, "must not be only digits") ||
				strings.Contains(tErr.Key, "must not be only lowercase letters") ||
				strings.Contains(tErr.Key, "must not be only uppercase letters") ||
				strings.Contains(tErr.Key, "must have at least 8 different glyphs") {
				api.Error(w, r, err, http.StatusBadRequest)
				return
			}
		}
		api.Error(w, r, t.Errorf("could not create user: %w", err), http.StatusInternalServerError)
		return
	}
	resp := userEntity.FilteredResponse(false)
	api.SuccessJson(w, r, resp)
}

// Me returns the current user
func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	userEntity, err := api.GetUserFromContext(r)
	if err != nil {
		api.Error(w, r, err, http.StatusUnauthorized)
		return
	}
	resp := userEntity.FilteredResponse(false)
	api.SuccessJson(w, r, resp)
}

// Get returns a specific user
func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userEntity, err := api.GetUserFromContext(r)
	if err != nil {
		api.Error(w, r, err, http.StatusUnauthorized)
		return
	}

	key := ps.ByName("key")
	isAdmin := api.IsAktivAdmin(*userEntity)

	if key != userEntity.Key && !isAdmin {
		api.Error(w, r, t.Errorf("Unauthorized access"), http.StatusForbidden)
		return
	}
	userEntityFetched, err := h.Service.DB.Users.Read(key, r.Context())
	if err != nil {
		api.Error(w, r, t.Errorf("user not found"), http.StatusNotFound)
		return
	}
	api.SuccessJson(w, r, userEntityFetched.FilteredResponse(isAdmin))
}

// List returns users dynamically filtered for admin
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	_, err := api.RequireAktivAdmin(r, h.Service.DB)
	if err != nil {
		api.Error(w, r, err, http.StatusUnauthorized)
		return
	}

	memStatus := r.URL.Query().Get("membership_status")
	hasClub := r.URL.Query().Get("has_club")

	users, err := h.Service.List(r.Context(), memStatus, hasClub)
	if err != nil {
		api.Error(w, r, err, http.StatusInternalServerError)
		return
	}

	var resp []*entities.User
	for _, u := range users {
		userCopy := u
		resp = append(resp, userCopy.FilteredResponse(false).(*entities.User))
	}
	if resp == nil {
		resp = make([]*entities.User, 0)
	}

	if r.Header.Get("Accept") == "text/csv" {
		api.SuccessCSV(w, r, "users_export.csv", users)
		return
	}

	api.SuccessJson(w, r, resp)
}

// Update allows a user to update their own fields, or an admin to update any user
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userEntity, err := api.GetUserFromContext(r)
	if err != nil {
		api.Error(w, r, err, http.StatusUnauthorized)
		return
	}

	key := ps.ByName("key")
	isAdmin := api.IsAktivAdmin(*userEntity)

	if key != userEntity.Key && !isAdmin {
		api.Error(w, r, t.Errorf("Unauthorized access"), http.StatusForbidden)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		api.Error(w, r, t.Errorf("invalid JSON body"), http.StatusBadRequest)
		return
	}

	for k, v := range updates {
		if s, ok := v.(string); ok {
			updates[k] = strings.TrimSpace(s)
		}
	}

	updatedUser, err := h.Service.UpdateUser(r.Context(), key, updates)
	if err != nil {
		api.Error(w, r, err, http.StatusBadRequest)
		return
	}

	api.SuccessJson(w, r, updatedUser.FilteredResponse(isAdmin))
}

// RequestEmailValidation - requires authentication
func (h *UserHandler) RequestEmailValidation(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	userEntity, err := api.GetUserFromContext(r)
	if err != nil {
		api.Error(w, r, err, http.StatusUnauthorized)
		return
	}

	var req struct {
		Email string `json:"email,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.Error(w, r, t.Errorf("read request body failed: %w", err), http.StatusBadRequest)
		return
	}

	err = h.Service.RequestEmailValidation(r.Context(), req.Email)
	if err != nil {
		api.Error(w, r, err, http.StatusBadRequest)
		return
	}

	targetEmail := req.Email
	if targetEmail == "" {
		targetEmail = userEntity.Email
	}

	api.SuccessJson(w, r, map[string]string{
		"message": t.T(t.Errorf("Validation email sent to %s", targetEmail), api.DetectLanguage(r)),
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

// RequestPasswordReset - public endpoint, requests password reset email
func (h *UserHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.Error(w, r, t.Errorf("read request body failed: %w", err), http.StatusBadRequest)
		return
	}
	if req.Email == "" {
		api.Error(w, r, t.Errorf("email must not be empty"), http.StatusBadRequest)
		return
	}
	err := h.Service.RequestPasswordReset(r.Context(), req.Email)
	if err != nil {
		api.Error(w, r, err, http.StatusBadRequest)
		return
	}
	api.SuccessJson(w, r, map[string]string{
		"message": t.T(t.Errorf("Password reset email sent to %s", req.Email), api.DetectLanguage(r)),
	})
}

// ShowResetPasswordForm - GET: show password reset HTML form
func (h *UserHandler) ShowResetPasswordForm(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	userKey := r.URL.Query().Get("key")
	expiryStr := r.URL.Query().Get("expiry")
	token := r.URL.Query().Get("token")

	if userKey == "" || expiryStr == "" || token == "" {
		api.Error(w, r, t.Errorf("missing required parameters"), http.StatusBadRequest)
		return
	}

	_, err := strconv.ParseInt(expiryStr, 10, 64)
	if err != nil {
		api.Error(w, r, t.Errorf("invalid expiry timestamp"), http.StatusBadRequest)
		return
	}

	// Show HTML form for password reset with JS for JSON POST
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="de">
<head>
    <meta charset="UTF-8">
    <title>Passwort zurücksetzen - DPV</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 400px; margin: 50px auto; padding: 20px; }
        input[type="password"], input[type="submit"] { width: 100%%; padding: 10px; margin: 8px 0; }
        .error { color: red; }
    </style>
</head>
<body>
    <h1>🔒 Passwort zurücksetzen</h1>
    <form id="resetForm">
        <input type="hidden" id="key" value="%s">
        <input type="hidden" id="expiry" value="%s">
        <input type="hidden" id="token" value="%s">
        <label for="password">Neues Passwort:</label>
        <input type="password" id="password" required>
        <label for="confirm">Passwort bestätigen:</label>
        <input type="password" id="confirm" required>
        <button type="submit">Passwort ändern</button>
    </form>
    <div id="result"></div>
    <script>
      document.getElementById('resetForm').onsubmit = async function(e) {
        e.preventDefault();
        const key = document.getElementById('key').value;
        const expiry = document.getElementById('expiry').value;
        const token = document.getElementById('token').value;
        const password = document.getElementById('password').value;
        const confirm = document.getElementById('confirm').value;
        const resultDiv = document.getElementById('result');
        resultDiv.textContent = '';
        const resp = await fetch(window.location.pathname, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ key, expiry, token, password, confirm })
        });
        const data = await resp.json();
        if (resp.ok) {
          resultDiv.textContent = '✅ Passwort erfolgreich geändert!';
        } else {
          resultDiv.textContent = 'Fehler: ' + (data.message || 'Unbekannter Fehler');
        }
      };
    </script>
</body>
</html>`, userKey, expiryStr, token)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	api.Success(w, r, []byte(html))
}

// HandleResetPassword - POST: handle password reset
func (h *UserHandler) HandleResetPassword(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req struct {
		Key      string `json:"key"`
		Expiry   string `json:"expiry"`
		Token    string `json:"token"`
		Password string `json:"password"`
		Confirm  string `json:"confirm"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.Error(w, r, t.Errorf("invalid JSON body"), http.StatusBadRequest)
		return
	}
	if req.Key == "" || req.Expiry == "" || req.Token == "" || req.Password == "" || req.Confirm == "" {
		api.Error(w, r, t.Errorf("missing required parameters"), http.StatusBadRequest)
		return
	}
	expiry, err := strconv.ParseInt(req.Expiry, 10, 64)
	if err != nil {
		api.Error(w, r, t.Errorf("invalid expiry timestamp"), http.StatusBadRequest)
		return
	}
	if req.Password != req.Confirm {
		api.Error(w, r, t.Errorf("passwords do not match"), http.StatusBadRequest)
		return
	}
	err = h.Service.ValidatePasswordReset(context.Background(), req.Key, expiry, req.Token, req.Password)
	if err != nil {
		api.Error(w, r, err, http.StatusBadRequest)
		return
	}
	api.SuccessJson(w, r, map[string]string{
		"message": t.T(t.Errorf("Password successfully changed"), api.DetectLanguage(r)),
	})
}

// UpdateRoles allows an admin to update a user's roles
func (h *UserHandler) UpdateRoles(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	_, err := api.RequireGlobalAdmin(r, h.Service.DB)
	if err != nil {
		api.Error(w, r, err, http.StatusUnauthorized)
		return
	}

	key := ps.ByName("key")
	var req struct {
		Roles []string `json:"roles"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.Error(w, r, t.Errorf("invalid JSON body"), http.StatusBadRequest)
		return
	}

	err = h.Service.UpdateRoles(r.Context(), key, req.Roles)
	if err != nil {
		api.Error(w, r, err, http.StatusBadRequest)
		return
	}

	// Fetch updated user to return
	updatedUser, err := h.Service.DB.Users.Read(key, r.Context())
	if err != nil {
		api.Error(w, r, t.Errorf("could not retrieve updated user"), http.StatusInternalServerError)
		return
	}

	api.SuccessJson(w, r, updatedUser.FilteredResponse(true))
}
