package handlers

import (
	"encoding/json"
	"eurovision-api/auth"
	"eurovision-api/db"
	"eurovision-api/models"
	"eurovision-api/utils"
	"net/http"
	"time"

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
	ranking.RankingID = GenerateShortID()

	err := db.CreateRanking(&ranking)

	if err != nil {
		logrus.Error("Error creating ranking: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *RankingHandler) DeleteRanking(w http.ResponseWriter, r *http.Request) {

}

/**
 * updates an existing user ranking
 */
func (h *RankingHandler) UpdateRanking(w http.ResponseWriter, r *http.Request) {

	ranking, valid := utils.DecodeRequestBody[models.UserRanking](w, r)

	if !valid {
		return
	}

	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// verify the ranking exists and belongs to the user
	existingRanking := getRanking(ranking.RankingID, w)
	if existingRanking == nil {
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

func getRanking(rankingID string, w http.ResponseWriter) *models.UserRanking {

	existingRanking, err := db.GetRankingByID(rankingID)

	if err != nil {
		if err.Error() == "ranking not found" {
			http.Error(w, "Ranking not found", http.StatusNotFound)
		} else {
			logrus.Error("Error fetching ranking: ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return nil
	}
	return existingRanking
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
