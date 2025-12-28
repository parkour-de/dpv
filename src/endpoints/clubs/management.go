package clubs

import (
	"dpv/dpv/src/api"
	"dpv/dpv/src/domain/entities"
	"dpv/dpv/src/repository/graph"
	"dpv/dpv/src/repository/t"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// List handles listing clubs with filters and pagination.
func (h *ClubHandler) List(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	user, ok := r.Context().Value("user").(*entities.User)
	if !ok || user == nil {
		api.Error(w, r, t.Errorf("user not found in context"), http.StatusUnauthorized)
		return
	}

	isAdmin := api.IsAdmin(*user)
	status := r.URL.Query().Get("status")
	skip, _ := api.ParseInt(r.URL.Query().Get("skip"))
	limit, _ := api.ParseInt(r.URL.Query().Get("limit"))

	var clubs []entities.Club
	var err error

	if isAdmin {
		options := graph.ClubQueryOptions{
			Skip:   skip,
			Limit:  limit,
			Status: status,
		}
		clubs, err = h.Service.GetAllClubs(r.Context(), options)
	} else {
		// Non-admins only see clubs they administer
		clubs, err = h.Service.ListClubs(r.Context(), user.Key)
		// Basic filtering for non-admins if status is provided
		if status != "" {
			var filtered []entities.Club
			for _, c := range clubs {
				if c.Membership.Status == status {
					filtered = append(filtered, c)
				}
			}
			clubs = filtered
		}
	}

	if err != nil {
		api.Error(w, r, err, http.StatusInternalServerError)
		return
	}

	var resp []entities.Club
	for _, c := range clubs {
		resp = append(resp, *FilteredResponse(&c))
	}

	api.SuccessJson(w, r, resp)
}
