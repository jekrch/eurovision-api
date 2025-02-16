package handlers

import (
	"encoding/json"
	"eurovision-api/auth"
	"eurovision-api/db"
	"eurovision-api/models"
	"eurovision-api/utils"
	"net/http"
	"time"

	"github.com/gorilla/mux"
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

	ranking, valid := utils.DecodeRequestBody[models.UserRanking](w, r)

	if !valid {
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

	// get rankingID from URL path variables
	vars := mux.Vars(r)
	rankingID := vars["rankingID"]

	if rankingID == "" {
		http.Error(w, "Ranking ID is required", http.StatusBadRequest)
		return
	}

	// Get the ranking using the existing helper function
	ranking := getAuthorizedRanking(w, r, rankingID, false)

	if ranking == nil {
		return
	}

	err := db.DeleteByFieldValue(db.RankingsIndex, "ranking_id", rankingID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Ranking deleted",
	})
}

/**
 * retrieves a ranking by ID from the URL path parameter
 */
func (h *RankingHandler) GetRanking(w http.ResponseWriter, r *http.Request) {

	// get rankingID from URL path variables
	vars := mux.Vars(r)
	rankingID := vars["rankingID"]

	if rankingID == "" {
		http.Error(w, "Ranking ID is required", http.StatusBadRequest)
		return
	}

	// get the ranking using the existing helper function
	ranking := getAuthorizedRanking(w, r, rankingID, true)

	if ranking == nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ranking)
}

/**
 * updates an existing user ranking
 */
func (h *RankingHandler) UpdateRanking(w http.ResponseWriter, r *http.Request) {

	ranking, valid := utils.DecodeRequestBody[models.UserRanking](w, r)

	if !valid {
		return
	}

	existingRanking := getAuthorizedRanking(w, r, ranking.RankingID, false)

	if existingRanking == nil {
		return
	}

	// preserve the original UserID and CreatedAt
	ranking.RankingID = existingRanking.RankingID
	ranking.UserID = existingRanking.UserID
	ranking.CreatedAt = existingRanking.CreatedAt
	ranking.UpdatedAt = time.Now()

	err := db.UpdateRanking(&ranking)

	if err != nil {
		logrus.Error("Error updating ranking: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

/*
retrieves an authorized user ranking by ID.
throw an error if
  - the ranking does not exist or
  - if the requesting user is not the owner of the ranking
  - if (allowPublic = true) then public rankings from other users are allowed
*/
func getAuthorizedRanking(w http.ResponseWriter, r *http.Request, rankingID string, allowPublic bool) *models.UserRanking {

	userID, err := auth.GetUserIDFromContext(r.Context())

	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return nil
	}

	existingRanking := getRanking(rankingID, w)

	if existingRanking == nil {
		return nil
	}

	if allowPublic && existingRanking.Public {
		return existingRanking
	}

	// if the requesting user is not the owner, throw unauth
	if existingRanking.UserID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return nil
	}

	return existingRanking
}

func getRanking(rankingID string, w http.ResponseWriter) *models.UserRanking {

	existingRanking, err := db.GetRankingByID(rankingID)

	if err != nil {
		if err.Error() == "ranking not found" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
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
