package middleware

import (
	"context"
	"dpv/dpv/src/api"
	"dpv/dpv/src/repository/graph"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func CORSMiddleware(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, x-altcha-spam-filter")
		// If Origin is present, set it. Otherwise allow all? No, allow origin from request is standard for dev/local.

		next(w, r, ps)
	}
}

func BasicAuthMiddleware(next httprouter.Handle, db *graph.Db) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		user, err := api.Authenticated(r, db)
		if err != nil {
			w.Header().Set("WWW-Authenticate", "Basic realm=DPV")
			http.Error(w, "Unauthorized", 401)
			return
		}

		// Store user in context for handlers
		ctx := context.WithValue(r.Context(), "user", user)
		next(w, r.WithContext(ctx), ps)
	}
}
