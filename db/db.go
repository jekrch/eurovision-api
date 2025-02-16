package db

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
)

const (
	usersIndex    = "users"
	RankingsIndex = "user_rankings"
	timeout       = 5 * time.Second
)

var (
	esClient *elastic.Client
	once     sync.Once
)

/**
 * initialize the es client and create the users index with mappings.
 */
func InitES() error {
	var initErr error
	once.Do(func() {
		client, err := elastic.NewClient(
			elastic.SetURL(os.Getenv("ELASTICSEARCH_URL")),
			elastic.SetSniff(false),
		)
		if err != nil {
			initErr = err
			return
		}
		esClient = client
		initErr = createUsersIndex()
		initErr = createRankingsIndex()
	})
	return initErr
}

/**
 * creates the provided index with the specified schema if it doesn't exist.
 */
func createIndex(indexName, schema string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	exists, err := esClient.IndexExists(indexName).Do(ctx)

	if err != nil {
		return fmt.Errorf("error checking index existence: %v", err)
	}

	if exists {
		return nil
	}

	createIndex, err := esClient.CreateIndex(indexName).Body(schema).Do(ctx)
	if err != nil {
		return fmt.Errorf("error creating index: %v", err)
	}
	if !createIndex.Acknowledged {
		return fmt.Errorf("index creation not acknowledged")
	}

	return nil
}

/**
 * creates a new document in the specified index.
 * Returns the response and any error that occurred.
 */
func Index(index string, body interface{}) (*elastic.IndexResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res, err := esClient.Index().
		Index(index).
		BodyJson(body).
		Refresh("true").
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("error indexing document: %v", err)
	}

	return res, nil
}

/**
 * gets the number of docs in the specified index.
 */
func Count(index string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	count, err := esClient.Count(index).Do(ctx)
	if err != nil {
		return 0, fmt.Errorf("error counting documents: %v", err)
	}

	return count, nil
}

func DeleteByFieldValue(indexName, fieldName, value string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	boolQuery := elastic.NewBoolQuery().
		Must(
			elastic.NewTermQuery(fieldName, value),
		)

	_, err := esClient.DeleteByQuery().
		Index(indexName).
		Query(boolQuery).
		Refresh("true").
		Do(ctx)

	if err != nil {
		logrus.Errorf("error deleting %s doc with %s = %s: %v", indexName, fieldName, value, err)
		return err
	}

	return nil
}

/*
counts the number of documents in the specified index where the field value matches
the provided value.
*/
func CountByFieldValue(indexName, fieldName, value string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	boolQuery := elastic.NewBoolQuery().
		Must(
			elastic.NewTermQuery(fieldName, value),
		)

	result, err := esClient.Count().
		Index(indexName).
		Query(boolQuery).
		Do(ctx)

	if err != nil {
		logrus.Errorf("error counting %s doc with %s = %s: %v", indexName, fieldName, value, err)
		return -1, err
	}

	return result, nil
}
