package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cloudnativego/gogo-engine"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

var (
	formatter = render.New(render.Options{
		IndentJSON: true,
	})
)

const (
	fakeMatchLocationResult = "/matches/5a003b78-409e-4452-b456-a6f0dcee05bd"
)

func TestCreateMatch(t *testing.T) {
	client := &http.Client{}
	repo := newInMemoryRepository()
	server := httptest.NewServer(
		http.HandlerFunc(createMatchHandler(formatter, repo)),
	)
	defer server.Close()

	body := []byte("{\"gridsize\": 19,\"playerWhite\": \"bob\",\"playerBlack\": \"alfred\"}")

	req, err := http.NewRequest("POST", server.URL, bytes.NewBuffer(body))

	if err != nil {
		t.Errorf("Error in creating POST request for createMatchHandler: %v", err)
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)

	if err != nil {
		t.Errorf("Error in POST to createMatchHandler: %v", err)
	}

	defer res.Body.Close()

	payload, err := ioutil.ReadAll(res.Body)

	if err != nil {
		t.Errorf("Error reading response body: %v", err)
	}

	if res.StatusCode != http.StatusCreated {
		t.Errorf("Expected response status 201, received %s", res.Status)
	}

	if _, ok := res.Header["Location"]; !ok {
		t.Errorf("Location Header is not set")
	}

	loc, headerOk := res.Header["Location"]
	if !headerOk {
		t.Error("Location header is not set.")
	} else {
		if !strings.Contains(loc[0], "/matches/") {
			t.Errorf("Location header should container '/matches/'")
		}
		if len(loc[0]) != len(fakeMatchLocationResult) {
			t.Errorf("Location value does not contain guid of new match")
		}
	}

	var matchResponse newMatchResponse
	err = json.Unmarshal(payload, &matchResponse)
	if err != nil {
		t.Errorf("Could not unmarshal payload into newMatchResponse object: %v", err)
	}

	if matchResponse.ID == "" || !strings.Contains(loc[0], matchResponse.ID) {
		t.Error("matchResponse.Id does not match Location header")
	}

	// After creating a match, match repository should have 1 item in it.
	matches, err := repo.getMatches()
	if err != nil {
		t.Errorf("Unexpected error in getMatches(): %s", err)
	}
	if len(matches) != 1 {
		t.Errorf("Expected a match repo of 1 match, got size %d", len(matches))
	}

	var match gogo.Match
	match = matches[0]
	if match.GridSize != matchResponse.GridSize {
		t.Errorf("Expected repo match and HTTP response gridsize to match. Got %d and %d", match.GridSize, matchResponse.GridSize)
	}

	if match.PlayerWhite != "bob" {
		t.Errorf("Repository match, white player should be bob, got %s", match.PlayerWhite)
	}

	if matchResponse.PlayerWhite != "bob" {
		t.Errorf("Expected white player to be bob, got %s", matchResponse.PlayerWhite)
	}

	if matchResponse.PlayerBlack != "alfred" {
		t.Errorf("Expected black player to be alfred, got %s", matchResponse.PlayerBlack)
	}

	fmt.Printf("Payload: %s", string(payload))
}

func TestCreateMatchRespondsToBadData(t *testing.T) {
	client := &http.Client{}
	repo := newInMemoryRepository()
	server := httptest.NewServer(http.HandlerFunc(createMatchHandler(formatter, repo)))
	defer server.Close()

	body1 := []byte("this is not valid json")
	body2 := []byte("{\"test\":\"this is valid json, but doesn't conform to server expectations.\"}")

	// Send invalid JSON
	req, err := http.NewRequest("POST", server.URL, bytes.NewBuffer(body1))
	if err != nil {
		t.Errorf("Error in creating POST request for createMatchHandler: %v", err)
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		t.Errorf("Error in POST to createMatchHandler: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Error("Sending invalid JSON should result in a bad request from server.")
	}

	req2, err2 := http.NewRequest("POST", server.URL, bytes.NewBuffer(body2))
	if err2 != nil {
		t.Errorf("Error in creating second POST request for invalid data on create match: %v", err2)
	}
	req2.Header.Add("Content-Type", "application/json")
	res2, _ := client.Do(req2)

	defer res2.Body.Close()
	if res2.StatusCode != http.StatusBadRequest {
		t.Error("Sending valid JSON but with incorrect or missing fields should result in a bad request and didn't.")
	}
}

func TestGetMatchListReturnsEmptyArrayForNoMatches(t *testing.T) {
	client := &http.Client{}
	repo := newInMemoryRepository()
	server := httptest.NewServer(http.HandlerFunc(getMatchListHandler(formatter, repo)))
	defer server.Close()
	req, _ := http.NewRequest("GET", server.URL, nil)

	resp, err := client.Do(req)

	if err != nil {
		t.Error("Errored when sending request to the server", err)
		return
	}

	defer resp.Body.Close()
	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error("Failed to read response from server", err)
	}

	var matchList []newMatchResponse
	err = json.Unmarshal(payload, &matchList)
	if err != nil {
		t.Errorf("Could not unmarshal payload into []newMatchResponse slice")
	}

	if len(matchList) != 0 {
		t.Errorf("Expected an empty list of match responses, got %d", len(matchList))
	}
}

func TestGetMatchListReturnsWhatsInRepository(t *testing.T) {
	client := &http.Client{}
	repo := newInMemoryRepository()
	repo.addMatch(gogo.NewMatch(19, "black", "white"))
	repo.addMatch(gogo.NewMatch(13, "bl", "wh"))
	repo.addMatch(gogo.NewMatch(19, "b", "w"))
	server := httptest.NewServer(http.HandlerFunc(getMatchListHandler(formatter, repo)))
	defer server.Close()
	req, _ := http.NewRequest("GET", server.URL, nil)

	resp, err := client.Do(req)

	if err != nil {
		t.Error("Errored when sending request to the server", err)
		return
	}

	defer resp.Body.Close()
	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error("Failed to read response from server", err)
	}

	var matchList []newMatchResponse
	err = json.Unmarshal(payload, &matchList)
	if err != nil {
		t.Errorf("Could not unmarshal payload into []newMatchResponse slice")
	}

	repoMatches, err := repo.getMatches()
	if err != nil {
		t.Errorf("Unexpected error in getMatches(): %s", err)
	}
	if len(matchList) != len(repoMatches) {
		t.Errorf("Match response size should have equaled repo size, sizes were: %d and %d", len(matchList), len(repoMatches))
	}

	for idx := 0; idx < 3; idx++ {
		if matchList[idx].GridSize != repoMatches[idx].GridSize {
			t.Errorf("Gridsize mismatch at index %d. Got %d and %d", idx, matchList[idx].GridSize, repoMatches[idx].GridSize)
		}
		if matchList[idx].PlayerBlack != matchList[idx].PlayerBlack {
			t.Errorf("PlayerBlack mismatch at index %d. Got %s and %s", idx, matchList[idx].PlayerBlack, repoMatches[idx].PlayerBlack)
		}
		if matchList[idx].PlayerWhite != matchList[idx].PlayerWhite {
			t.Errorf("PlayerWhite mismatch at index %d. Got %s and %s", idx, matchList[idx].PlayerWhite, repoMatches[idx].PlayerWhite)
		}
	}
}

func TestGetMatchDetailsReturnsExistingMatch(t *testing.T) {
	var (
		request  *http.Request
		recorder *httptest.ResponseRecorder
	)

	repo := newInMemoryRepository()
	server := MakeTestServer(repo)

	targetMatch := gogo.NewMatch(19, "black", "white")
	repo.addMatch(targetMatch)
	targetMatchID := targetMatch.ID

	recorder = httptest.NewRecorder()
	request, _ = http.NewRequest("GET", "/matches/"+targetMatchID, nil)
	server.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected %v; received %v", http.StatusOK, recorder.Code)
	}

	var match matchDetailsResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &match)
	if err != nil {
		t.Errorf("Error unmarshaling match details: %s", err)
	}
	if match.GridSize != 19 {
		t.Errorf("Expected match gridsize to be 19; received %d", match.GridSize)
	}
}

func MakeTestServer(repository matchRepository) *negroni.Negroni {
	server := negroni.New() // don't need all the middleware here or logging.
	mx := mux.NewRouter()
	initRoutes(mx, formatter, repository)
	server.UseHandler(mx)
	return server
}
