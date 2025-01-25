package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type Vote struct {
	VoteString string    `json:"vote_string"`
	IP         string    `json:"ip"`
	Location   string    `json:"location"`
	Year       int       `json:"year"`
	Timestamp  time.Time `json:"timestamp"`
}

var esClient *elasticsearch.Client

func main() {
	// initialize elasticsearch client
	cfg := elasticsearch.Config{
		Addresses: []string{os.Getenv("ELASTICSEARCH_URL")},
	}

	var err error
	esClient, err = elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error creating Elasticsearch client: %s", err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/vote", handleVote).Methods("POST")
	r.HandleFunc("/votes/count", getVoteCount).Methods("GET")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func handleVote(w http.ResponseWriter, r *http.Request) {
	var vote Vote
	if err := json.NewDecoder(r.Body).Decode(&vote); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set IP and timestamp
	vote.IP = r.RemoteAddr
	vote.Timestamp = time.Now()

	log.Printf("received vote from %s", vote.IP)

	// Index the vote in Elasticsearch
	voteJSON, err := json.Marshal(vote)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = esClient.Index(
		"eurovision_votes",
		strings.NewReader(string(voteJSON)),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func getVoteCount(w http.ResponseWriter, r *http.Request) {
	res, err := esClient.Count(
		esClient.Count.WithIndex("eurovision_votes"),
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
