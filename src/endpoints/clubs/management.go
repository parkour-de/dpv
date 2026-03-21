package clubs

import (
	"context"
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

	status := r.URL.Query().Get("status")
	skip, _ := api.ParseInt(r.URL.Query().Get("skip"))
	limit, _ := api.ParseInt(r.URL.Query().Get("limit"))
	missingCensusYear, _ := api.ParseInt(r.URL.Query().Get("missing_census_year"))

	clubs, err := h.getClubsForUser(r.Context(), user, status, skip, limit, missingCensusYear)
	if err != nil {
		api.Error(w, r, err, http.StatusInternalServerError)
		return
	}

	if r.Header.Get("Accept") == "text/csv" {
		api.SuccessCSV(w, r, "clubs_export.csv", clubs)
		return
	}

	resp := h.buildFilteredClubResponse(clubs)
	api.SuccessJson(w, r, resp)
}

func (h *ClubHandler) getClubsForUser(ctx context.Context, user *entities.User, status string, skip, limit, missingCensusYear int) ([]entities.Club, error) {
	if api.IsAdmin(*user) {
		options := graph.ClubQueryOptions{
			Skip:              skip,
			Limit:             limit,
			Status:            status,
			MissingCensusYear: missingCensusYear,
		}
		return h.Service.GetAllClubs(ctx, options)
	}

	// Non-admins only see clubs they administer
	clubs, err := h.Service.ListClubs(ctx, user.Key)
	if err != nil {
		return nil, err
	}

	if status != "" {
		return h.filterClubsByStatus(clubs, status), nil
	}
	return clubs, nil
}

func (h *ClubHandler) filterClubsByStatus(clubs []entities.Club, status string) []entities.Club {
	var filtered []entities.Club
	for _, c := range clubs {
		if c.Membership.Status == status {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

func (h *ClubHandler) buildFilteredClubResponse(clubs []entities.Club) []entities.Club {
	var resp []entities.Club
	for _, c := range clubs {
		resp = append(resp, *(c.FilteredResponse().(*entities.Club)))
	}
	return resp
}
