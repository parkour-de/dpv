package clubs

import (
	"dpv/dpv/src/api"
	"dpv/dpv/src/domain/entities"
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

	currentYear := time.Now().Year()
	census, _ := h.Service.DB.GetCensus(r.Context(), key, currentYear)

	queryName := h.getQueryName(club.Name)
	users, err := h.Service.DB.GetUsersByYourClub(r.Context(), queryName)
	if err != nil {
		api.Error(w, r, t.Errorf("failed to search portal users: %w", err), http.StatusInternalServerError)
		return
	}

	resp := activeMembersResponse{
		ExactMatches:   make([]activeMemberMatch, 0),
		PartialMatches: make([]activeMemberMatch, 0),
	}

	for _, u := range users {
		match := h.findMatchInCensus(u, census)
		if match.MatchType == "exact" {
			resp.ExactMatches = append(resp.ExactMatches, match)
		} else {
			resp.PartialMatches = append(resp.PartialMatches, match)
		}
	}

	api.SuccessJson(w, r, resp)
}

func (h *ClubHandler) getQueryName(clubName string) string {
	if len(clubName) > 4 && strings.Contains(strings.ToLower(clubName), " e.v.") {
		return clubName[:len(clubName)-5]
	}
	return clubName
}

func (h *ClubHandler) findMatchInCensus(u entities.User, census *entities.Census) activeMemberMatch {
	userNormFirst := h.normalizeString(u.FirstName)
	userNormLast := h.normalizeString(u.LastName)
	userDOB := h.formatDOB(u.DateOfBirth)

	if census != nil {
		for cIdx := range census.Members {
			cEntry := &census.Members[cIdx]
			if cEntry.Firstname == "" && cEntry.Lastname == "" {
				continue
			}

			censusNormFirst := h.normalizeString(cEntry.Firstname)
			censusNormLast := h.normalizeString(cEntry.Lastname)
			censusDOB := h.formatDOB(h.normalizeString(cEntry.BirthDate))

			if userNormFirst == censusNormFirst && userNormLast == censusNormLast {
				match := activeMemberMatch{
					User:           u.FilteredResponse(),
					Source:         "census_and_profile",
					CensusName:     cEntry.Firstname + " " + cEntry.Lastname,
					CensusDOB:      cEntry.BirthDate,
					PortalName:     u.FirstName + " " + u.LastName,
					PortalDOB:      u.DateOfBirth,
					PortalYourClub: u.YourClub,
				}

				if userDOB == censusDOB {
					match.MatchType = "exact"
					cEntry.Firstname = ""
					cEntry.Lastname = ""
					cEntry.BirthDate = ""
					return match
				}

				match.MatchType = "partial_name_match"
				return match
			}
		}
	}

	return activeMemberMatch{
		User:           u.FilteredResponse(),
		Source:         "profile_only",
		PortalName:     u.FirstName + " " + u.LastName,
		PortalDOB:      u.DateOfBirth,
		PortalYourClub: u.YourClub,
		MatchType:      "dein_verein_only",
	}
}

func (h *ClubHandler) normalizeString(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func (h *ClubHandler) formatDOB(dob string) string {
	if strings.Contains(dob, "T") {
		return strings.Split(dob, "T")[0]
	}
	return dob
}
