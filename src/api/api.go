package api

import (
	"dpv/dpv/src/domain/entities"
	"dpv/dpv/src/repository/graph"
	"dpv/dpv/src/repository/security"
	"dpv/dpv/src/repository/t"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func GetUserFromContext(r *http.Request) (*entities.User, error) {
	user, ok := r.Context().Value("user").(*entities.User)
	if !ok || user == nil {
		return nil, t.Errorf("user not found in context")
	}
	return user, nil
}

func Authenticated(r *http.Request, db *graph.Db) (*entities.User, error) {
	email, password, ok := r.BasicAuth()
	if !ok {
		return nil, t.Errorf("authorization header missing or not using Basic Auth")
	}
	users, err := db.GetUsersByEmail(r.Context(), email)
	if err != nil || len(users) != 1 {
		return nil, t.Errorf("user not found or multiple users returned")
	}
	user := users[0]
	authenticated := security.CheckPasswordHash(user.PasswordHash, password)
	if !authenticated {
		return nil, t.Errorf("invalid credentials")
	}
	return &user, nil
}

func IsAdmin(user entities.User) bool {
	return contains(user.Roles, "admin")
}

func contains(roles []string, s string) bool {
	for _, role := range roles {
		if role == s {
			return true
		}
	}
	return false
}

func RequireGlobalAdmin(r *http.Request, db *graph.Db) (*entities.User, error) {
	user, err := Authenticated(r, db)
	if err != nil {
		return nil, t.Errorf("authentication failed: %w", err)
	}
	if !IsAdmin(*user) {
		return nil, t.Errorf("you are not an administrator")
	}
	return user, nil
}

func SuccessJson(w http.ResponseWriter, r *http.Request, data interface{}) {
	jsonMsg, err := json.Marshal(data)
	if err != nil {
		Error(w, r, t.Errorf("serialising response failed: %w", err), 400)
		return
	} else {
		w.Header().Set("Content-Type", "application/json")
		Success(w, r, jsonMsg)
	}
}

func Success(w http.ResponseWriter, r *http.Request, jsonMsg []byte) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if _, err := w.Write(jsonMsg); err != nil {
		log.Printf("Error writing response: %v", err)
	}

	log.Printf(
		"%s %s %s 200",
		r.Method,
		r.RequestURI,
		r.RemoteAddr,
	)
}

func Error(w http.ResponseWriter, r *http.Request, err error, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	if err == nil {
		err = t.Errorf("nil err")
	}
	logErr := err
	errorMsgJSON, err := json.Marshal(ErrorResponse{
		err.Error(),
	})
	if err != nil {
		log.Println(err)
	} else {
		if _, err = w.Write(errorMsgJSON); err != nil {
			log.Printf("Error writing response: %v", err)
		}
	}

	log.Printf(
		"%s %s %s %d %s",
		r.Method,
		r.RequestURI,
		r.RemoteAddr,
		code,
		logErr.Error(),
	)
}

func MakeSet(queryParam string) map[string]struct{} {
	set := make(map[string]struct{})
	if queryParam != "" {
		tokens := strings.Split(queryParam, ",")
		for _, token := range tokens {
			set[token] = struct{}{}
		}
	}
	return set
}

func ParseInt(queryValue string) (int, error) {
	if queryValue == "" {
		return 0, nil
	}
	return strconv.Atoi(queryValue)
}

func ParseFloat(queryValue string) (float64, error) {
	if queryValue == "" {
		return 0, nil
	}
	return strconv.ParseFloat(queryValue, 64)
}
