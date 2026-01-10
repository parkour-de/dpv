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

	"golang.org/x/text/language"
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

func DetectLanguage(r *http.Request) string {
	targetLang := "en" // Default

	// Check Context (User profile)
	if user, ok := r.Context().Value("user").(*entities.User); ok && user.Language != "" {
		targetLang = user.Language
	} else if custom := r.Header.Get("X-Language"); custom != "" {
		// Check Custom Header (e.g. forced by frontend)
		targetLang = custom
	} else {
		// Check Accept-Language
		acceptLang := r.Header.Get("Accept-Language")
		tags, _, _ := language.ParseAcceptLanguage(acceptLang)
		if len(tags) > 0 {
			targetLang = tags[0].String()
		}
	}
	return targetLang
}

func Error(w http.ResponseWriter, r *http.Request, err error, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)

	if err == nil {
		err = t.Errorf("nil err")
	}

	// 1. Determine Language
	targetLang := DetectLanguage(r)

	// 2. Load the map for that language
	langMap := t.GetMapFor(targetLang)

	// 3. Translate the error tree
	finalMsg := t.Translate(err, langMap)

	logErr := err
	errorMsgJSON, marshalErr := json.Marshal(ErrorResponse{
		Message: finalMsg,
	})
	if marshalErr != nil {
		log.Println(marshalErr)
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

func SanitizeFilename(name string) string {
	// Simple cleanup: strictly alphanumeric, hyphen, underscore
	// Everything else becomes an underscore
	safe := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			return r
		}
		return '_'
	}, name)
	// clean up multiple underscores or leading/trailing
	return strings.Trim(safe, "_")
}
