package middleware

import (
	"context"
	"dpv/dpv/src/api"
	"dpv/dpv/src/repository/graph"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

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
