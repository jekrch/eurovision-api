package handlers

import (
	"encoding/json"

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

/**
 * Processes vote and stores in eurovision_votes index
 */
func (h *VoteHandler) HandleVote(w http.ResponseWriter, r *http.Request) {
	var vote models.Vote
	if err := json.NewDecoder(r.Body).Decode(&vote); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	vote.Timestamp = time.Now()

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

/**
 * Returns the number of votes cast
 */
func (h *VoteHandler) GetVoteCount(w http.ResponseWriter, r *http.Request) {
	res, err := db.Count(
		"eurovision_votes",
	)
	if err != nil {
		logrus.Error("An error occurred while fetching vote count: ", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
