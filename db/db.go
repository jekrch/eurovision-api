package db

import (
	"context"
	"encoding/json"
	"eurovision-api/models"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/olivere/elastic/v7"
)

const (
	usersIndex = "users"
	timeout    = 5 * time.Second
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
	})
	return initErr
}

/**
 * creates the users index with proper mappings if it doesn't exist.
 */
func createUsersIndex() error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	exists, err := esClient.IndexExists(usersIndex).Do(ctx)
	if err != nil {
		return fmt.Errorf("error checking index existence: %v", err)
	}

	if exists {
		return nil
	}

	mapping := `{
		"mappings": {
			"properties": {
				"email": {
					"type": "keyword"
				},
				"password_hash": {
					"type": "keyword"
				},
				"confirmed": {
					"type": "boolean"
				},
				"confirmation_token": {
					"type": "keyword"
				},
				"token_expiry": {
					"type": "date"
				},
				"created_at": {
					"type": "date"
				}
			}
		}
	}`

	createIndex, err := esClient.CreateIndex(usersIndex).Body(mapping).Do(ctx)
	if err != nil {
		return fmt.Errorf("error creating index: %v", err)
	}
	if !createIndex.Acknowledged {
		return fmt.Errorf("index creation not acknowledged")
	}

	return nil
}

/**
 * checks if an email is already registered in the users index
 */
func EmailExists(email string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	query := elastic.NewTermQuery("email", email)
	count, err := esClient.Count(usersIndex).Query(query).Do(ctx)
	if err != nil {
		return false, fmt.Errorf("error checking email: %v", err)
	}

	return count > 0, nil
}

/**
 * creates a new user in the Elasticsearch users index
 */
func CreateUser(user *models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := esClient.Index().
		Index(usersIndex).
		BodyJson(user).
		Refresh("true").
		Do(ctx)

	if err != nil {
		return fmt.Errorf("error creating user: %v", err)
	}

	return nil
}

/**
 * gets a user by their confirmation token
 */
func GetUserByToken(token string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	query := elastic.NewTermQuery("confirmation_token", token)
	result, err := esClient.Search().
		Index(usersIndex).
		Query(query).
		Size(1).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("error getting user: %v", err)
	}

	if result.TotalHits() == 0 {
		return nil, fmt.Errorf("user not found")
	}

	var user models.User
	err = json.Unmarshal(result.Hits.Hits[0].Source, &user)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling user: %v", err)
	}

	return &user, nil
}

/**
 * gets a user by their email address.
 */
func GetUserByEmail(email string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	query := elastic.NewTermQuery("email", email)
	result, err := esClient.Search().
		Index(usersIndex).
		Query(query).
		Size(1).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("error getting user: %v", err)
	}

	if result.TotalHits() == 0 {
		return nil, fmt.Errorf("user not found")
	}

	var user models.User
	err = json.Unmarshal(result.Hits.Hits[0].Source, &user)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling user: %v", err)
	}

	return &user, nil
}

/**
 * updates the user's confirmed status and removes the confirmation token.
 * Returns an error if the user is not found or the operation fails.
 */
func ConfirmUser(email string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	script := elastic.NewScript("ctx._source.confirmed = true; ctx._source.confirmation_token = null")
	query := elastic.NewTermQuery("email", email)

	result, err := esClient.UpdateByQuery(usersIndex).
		Query(query).
		Script(script).
		Refresh("true").
		Do(ctx)

	if err != nil {
		return fmt.Errorf("error confirming user: %v", err)
	}

	if result.Updated == 0 {
		return fmt.Errorf("user not found with email: %s", email)
	}

	return nil
}

/**
 * removes all unconfirmed users created before the cutoff time.
 * Returns an error if the operation fails.
 */
func DeleteUnconfirmedUsers(cutoff time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	boolQuery := elastic.NewBoolQuery().
		Must(
			elastic.NewTermQuery("confirmed", false),
			elastic.NewRangeQuery("created_at").Lt(cutoff),
		)

	_, err := esClient.DeleteByQuery().
		Index(usersIndex).
		Query(boolQuery).
		Refresh("true").
		Do(ctx)

	if err != nil {
		return fmt.Errorf("error deleting unconfirmed users: %v", err)
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
