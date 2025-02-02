package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"eurovision-api/db"
	"eurovision-api/models"
	"strings"

	"github.com/sirupsen/logrus"
)

type VoteHandler struct {
}

func NewVoteHandler() *VoteHandler {
	return &VoteHandler{}
}

func (h *VoteHandler) HandleVote(w http.ResponseWriter, r *http.Request) {
	var vote models.Vote
	if err := json.NewDecoder(r.Body).Decode(&vote); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// set IP and timestamp
	vote.IP = r.RemoteAddr
	vote.Timestamp = time.Now()

	// get location from IP
	assignLocationFromIP(&vote)

	logrus.Printf("received vote from %s; location: %v", vote.IP, vote.Location)

	// index the vote in Elasticsearch
	voteJSON, err := json.Marshal(vote)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = db.Index(
		"eurovision_votes",
		strings.NewReader(string(voteJSON)),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func assignLocationFromIP(vote *models.Vote) {

	locationURL := fmt.Sprintf("https://ipapi.co/%s/json/", vote.IP)
	resp, err := http.Get(locationURL)
	if err != nil {
		log.Printf("Error getting location: %v", err)
	} else {
		var ipLoc models.IPLocation
		if err := json.NewDecoder(resp.Body).Decode(&ipLoc); err != nil {
			log.Printf("Error decoding location: %v", err)
		} else {
			vote.Location = ipLoc
			vote.Country = ipLoc.CountryName
		}
		resp.Body.Close()
	}
}

func (h *VoteHandler) GetVoteCount(w http.ResponseWriter, r *http.Request) {
	res, err := db.Count(
		"eurovision_votes",
	)
	if err != nil {
		logrus.Error("An error occurred while fetching vote count: ", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var count struct {
		Count int `json:"count`
	}

	if err := json.NewDecoder(res.Body).Decode(&count); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(count)
}
