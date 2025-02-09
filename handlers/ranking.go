package handlers

import (
	"encoding/json"
	"eurovision-api/auth"
	"eurovision-api/db"
	"eurovision-api/models"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type RankingHandler struct {
}

func NewRankingHandler() *RankingHandler {
	return &RankingHandler{}
}

/**
 * creates a new user ranking
 */

func (h *RankingHandler) CreateRanking(w http.ResponseWriter, r *http.Request) {
	var ranking models.UserRanking
	if err := json.NewDecoder(r.Body).Decode(&ranking); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID, nil := auth.GetUserIDFromContext(r.Context())

	ranking.CreatedAt = time.Now()
	ranking.UserID = userID
	ranking.RankingID = uuid.New().String()

	err := db.CreateRanking(&ranking)

	if err != nil {
		logrus.Error("Error creating ranking: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

/**
 * updates an existing user ranking
 */
func (h *RankingHandler) UpdateRanking(w http.ResponseWriter, r *http.Request) {
	var ranking models.UserRanking
	if err := json.NewDecoder(r.Body).Decode(&ranking); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// verify the ranking exists and belongs to the user
	existingRanking, err := db.GetRankingByID(ranking.RankingID)

	if err != nil {
		if err.Error() == "ranking not found" {
			http.Error(w, "Ranking not found", http.StatusNotFound)
		} else {
			logrus.Error("Error fetching ranking: ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Check if the ranking belongs to the requesting user
	if existingRanking.UserID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Preserve the original UserID and CreatedAt
	ranking.UserID = existingRanking.UserID
	ranking.CreatedAt = existingRanking.CreatedAt
	ranking.UpdatedAt = time.Now()

	err = db.UpdateRanking(&ranking)

	if err != nil {
		logrus.Error("Error updating ranking: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

/**
 * Retrieves all rankings for a specific user
 */
func (h *RankingHandler) GetUserRankings(w http.ResponseWriter, r *http.Request) {

	userID, nil := auth.GetUserIDFromContext(r.Context())

	if userID == "" {
		http.Error(w, "userId is required", http.StatusBadRequest)
		return
	}

	rankings, err := db.GetRankingsByUserID(userID)
	if err != nil {
		logrus.Error("Error fetching rankings: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rankings)
}
