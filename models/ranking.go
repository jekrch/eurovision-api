package models

import (
	"time"
)

type UserRanking struct {
	UserID      string    `json:"user_id"`
	RankingID   string    `json:"ranking_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Year        int       `json:"year"`
	Ranking     string    `json:"ranking"`
	GroupIDs    []string  `json:"group_ids"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"created_at"`
}
