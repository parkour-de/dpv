package clubs

import (
	"dpv/dpv/src/api"
	"dpv/dpv/src/repository/t"
	"net/http"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
)

type activeMemberMatch struct {
	User           interface{} `json:"user,omitempty"`
	Source         string      `json:"source"`
	CensusName     string      `json:"census_name,omitempty"`
	CensusDOB      string      `json:"census_dob,omitempty"`
	PortalName     string      `json:"portal_name,omitempty"`
	PortalDOB      string      `json:"portal_dob,omitempty"`
	PortalYourClub string      `json:"portal_your_club,omitempty"`
	MatchType      string      `json:"match_type"`
}

type activeMembersResponse struct {
	ExactMatches   []activeMemberMatch `json:"exact_matches"`
	PartialMatches []activeMemberMatch `json:"partial_matches"`
}

// GetActiveMembers strictly returns the users who have placed the club in their "your club" and pairs them against the census
func (h *ClubHandler) GetActiveMembers(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Either an admin or an aktivadmin can view this list
	_, err := api.RequireAktivAdmin(r, h.Service.DB)
	if err != nil {
		api.Error(w, r, err, http.StatusUnauthorized)
		return
	}

	key := ps.ByName("key")
	club, err := h.Service.DB.GetClubByKey(r.Context(), key)
	if err != nil {
		api.Error(w, r, t.Errorf("failed to get club: %w", err), http.StatusNotFound)
		return
	}

	// 1. Fetch current year's census
	currentYear := time.Now().Year()
	census, _ := h.Service.DB.GetCensus(r.Context(), key, currentYear)

	// 2. Fetch users who have selected this club
	// If club name is long, might just use a significant part of it, but full name is fine for now
	queryName := club.Name
	if len(queryName) > 4 && strings.Contains(strings.ToLower(queryName), " e.v.") {
		queryName = queryName[:len(queryName)-5]
	}

	users, err := h.Service.DB.GetUsersByYourClub(r.Context(), queryName)
	if err != nil {
		api.Error(w, r, t.Errorf("failed to search portal users: %w", err), http.StatusInternalServerError)
		return
	}

	resp := activeMembersResponse{
		ExactMatches:   make([]activeMemberMatch, 0),
		PartialMatches: make([]activeMemberMatch, 0),
	}

	// Helper to normalize strings for comparison
	normalize := func(s string) string {
		return strings.ToLower(strings.TrimSpace(s))
	}

	// Helper to format DOB from user Entity to match census "YYYY-MM-DD"
	formatDOB := func(dob string) string {
		if strings.Contains(dob, "T") {
			return strings.Split(dob, "T")[0]
		}
		return dob
	}

	// Process all users found by "Dein Verein"
	for _, u := range users {
		userNormFirst := normalize(u.FirstName)
		userNormLast := normalize(u.LastName)
		userDOB := formatDOB(u.DateOfBirth)

		matchFound := false

		if census != nil {
			for cIdx := range census.Members {
				cEntry := &census.Members[cIdx]
				censusNormFirst := normalize(cEntry.Firstname)
				censusNormLast := normalize(cEntry.Lastname)
				censusDOB := formatDOB(normalize(cEntry.BirthDate))

				// Exact match: First Name + Last Name + DOB
				if userNormFirst == censusNormFirst && userNormLast == censusNormLast && userDOB == censusDOB {
					resp.ExactMatches = append(resp.ExactMatches, activeMemberMatch{
						User:           u.FilteredResponse(true),
						Source:         "census_and_profile",
						CensusName:     cEntry.Firstname + " " + cEntry.Lastname,
						CensusDOB:      cEntry.BirthDate,
						PortalName:     u.FirstName + " " + u.LastName,
						PortalDOB:      u.DateOfBirth,
						PortalYourClub: u.YourClub,
						MatchType:      "exact",
					})
					matchFound = true
					// Empty the entry so we don't match it again if we process censuses later
					cEntry.Firstname = ""
					cEntry.Lastname = ""
					cEntry.BirthDate = ""
					break
				}

				// Partial match: First Name + Last Name (but different DOB)
				if userNormFirst == censusNormFirst && userNormLast == censusNormLast {
					resp.PartialMatches = append(resp.PartialMatches, activeMemberMatch{
						User:           u.FilteredResponse(true),
						Source:         "census_and_profile",
						CensusName:     cEntry.Firstname + " " + cEntry.Lastname,
						CensusDOB:      cEntry.BirthDate,
						PortalName:     u.FirstName + " " + u.LastName,
						PortalDOB:      u.DateOfBirth,
						PortalYourClub: u.YourClub,
						MatchType:      "partial_name_match",
					})
					matchFound = true
					break
				}
			}
		}

		if !matchFound {
			// They wrote "Dein Verein" but didn't match anything in the census
			resp.PartialMatches = append(resp.PartialMatches, activeMemberMatch{
				User:           u.FilteredResponse(true),
				Source:         "profile_only",
				PortalName:     u.FirstName + " " + u.LastName,
				PortalDOB:      u.DateOfBirth,
				PortalYourClub: u.YourClub,
				MatchType:      "dein_verein_only",
			})
		}
	}

	api.SuccessJson(w, r, resp)
}
