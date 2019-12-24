package service

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/cloudnativego/gogo-engine"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

func createMatchHandler(formatter *render.Render, repo matchRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		payload, _ := ioutil.ReadAll(req.Body)
		var newMatchRequest newMatchRequest
		err := json.Unmarshal(payload, &newMatchRequest)
		if err != nil {
			formatter.Text(w, http.StatusBadRequest, "Failed to parse match request")
			return
		}
		if !newMatchRequest.isValid() {
			formatter.Text(w, http.StatusBadRequest, "Invalid new match request")
			return
		}

		newMatch := gogo.NewMatch(newMatchRequest.GridSize, newMatchRequest.PlayerBlack, newMatchRequest.PlayerWhite)
		repo.addMatch(newMatch)
		var mr newMatchResponse
		mr.copyMatch(newMatch)
		w.Header().Add("Location", "/matches/"+newMatch.ID)
		formatter.JSON(w, http.StatusCreated, &mr)
	}
}

func getMatchListHandler(formatter *render.Render, repo matchRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		repoMatches, err := repo.getMatches()
		if err == nil {
			matches := make([]newMatchResponse, len(repoMatches))
			for idx, match := range repoMatches {
				matches[idx].copyMatch(match)
			}
			formatter.JSON(w, http.StatusOK, matches)
		} else {
			formatter.JSON(w, http.StatusNotFound, err.Error())
		}

	}
}

func getMatchDetailsHandler(formatter *render.Render, repo matchRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		matchID := vars["id"]
		repoMatch, err := repo.getMatch(matchID)
		if err != nil {
			formatter.JSON(w, http.StatusNotFound, err.Error())
		} else {
			var mdr matchDetailsResponse
			mdr.copyMatch(repoMatch)
			formatter.JSON(w, http.StatusOK, &mdr)
		}
	}
}
