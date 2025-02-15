package db

import (
	"context"
	"encoding/json"
	"eurovision-api/models"
	"fmt"

	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
)

/**
 * creates the rankings index with proper mappings if it doesn't exist.
 */
func createRankingsIndex() error {

	mapping := `{
        "mappings": {
            "properties": {
                "user_id": {
                    "type": "keyword"
                },
                "ranking_id": {
                    "type": "keyword"
                },
                "name": {
                    "type": "text"
                },
                "description": {
                    "type": "text"
                },
                "year": {
                    "type": "integer"
                },
                "ranking": {
                    "type": "keyword"
                },
                "group_ids": {
                    "type": "keyword"
                },
                "created_at": {
                    "type": "date"
                }
            }
        }
    }`

	return createIndex(rankingsIndex, mapping)
}

/**
 * creates a new ranking in the user_rankings index
 */
func CreateRanking(ranking *models.UserRanking) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := esClient.Index().
		Index(rankingsIndex).
		BodyJson(ranking).
		Refresh("true").
		Do(ctx)

	if err != nil {
		return fmt.Errorf("error creating ranking: %v", err)
	}

	return nil
}

/**
 * gets all rankings for a specific user by ID
 */
func GetRankingsByUserID(userID string) ([]models.UserRanking, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	query := elastic.NewTermQuery("user_id", userID)
	result, err := esClient.Search().
		Index(rankingsIndex).
		Query(query).
		Sort("created_at", false).
		Size(100).
		Do(ctx)

	logrus.Print("query: ", query)
	logrus.Print("result: ", result)

	if err != nil {
		return nil, fmt.Errorf("error getting rankings: %v", err)
	}

	var rankings []models.UserRanking
	for _, hit := range result.Hits.Hits {
		var ranking models.UserRanking
		err := json.Unmarshal(hit.Source, &ranking)
		if err != nil {
			return nil, fmt.Errorf("error unmarshaling ranking: %v", err)
		}
		rankings = append(rankings, ranking)
	}

	return rankings, nil
}

/**
 * gets a ranking by its ID
 */
func GetRankingByID(rankingID string) (*models.UserRanking, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	query := elastic.NewTermQuery("ranking_id", rankingID)
	result, err := esClient.Search().
		Index(rankingsIndex).
		Query(query).
		Size(1).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("error getting ranking: %v", err)
	}

	if result.TotalHits() == 0 {
		return nil, fmt.Errorf("ranking not found")
	}

	var ranking models.UserRanking
	err = json.Unmarshal(result.Hits.Hits[0].Source, &ranking)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling ranking: %v", err)
	}

	return &ranking, nil
}

/**
 * updates an existing ranking in the user_rankings index
 */
func UpdateRanking(ranking *models.UserRanking) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := esClient.Update().
		Index(rankingsIndex).
		Id(ranking.RankingID).
		Doc(ranking).
		Refresh("true").
		Do(ctx)

	if err != nil {
		return fmt.Errorf("error updating ranking: %v", err)
	}

	return nil
}
