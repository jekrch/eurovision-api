package db

import (
	"context"
	"encoding/json"
	"eurovision-api/models"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

const (
	usersIndex = "users"
	timeout    = 5 * time.Second
)

var (
	esClient *elasticsearch.Client
	once     sync.Once
)

// InitES initializes the Elasticsearch client and creates necessary indices
func InitES() error {
	var initErr error
	once.Do(func() {
		cfg := elasticsearch.Config{
			Addresses: []string{os.Getenv("ELASTICSEARCH_URL")},
		}

		esClient, initErr = elasticsearch.NewClient(cfg)
		if initErr != nil {
			return
		}

		// Create users index with proper mappings
		initErr = createUsersIndex()
	})
	return initErr
}

func createUsersIndex() error {
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

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res, err := esClient.Indices.Create(
		usersIndex,
		esClient.Indices.Create.WithBody(strings.NewReader(mapping)),
		esClient.Indices.Create.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("error creating index: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		// Ignore error if index already exists
		if strings.Contains(res.String(), "resource_already_exists_exception") {
			return nil
		}
		return fmt.Errorf("error creating index: %s", res.String())
	}

	return nil
}

// EmailExists checks if an email is already registered
func EmailExists(email string) (bool, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				"email": email,
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res, err := esClient.Search(
		esClient.Search.WithIndex(usersIndex),
		esClient.Search.WithBody(strings.NewReader(mustToJSON(query))),
		esClient.Search.WithContext(ctx),
	)
	if err != nil {
		return false, fmt.Errorf("error checking email: %v", err)
	}
	defer res.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("error parsing response: %v", err)
	}

	hits := result["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)
	return hits > 0, nil
}

// CreateUser creates a new user in Elasticsearch
func CreateUser(user *models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res, err := esClient.Index(
		usersIndex,
		strings.NewReader(mustToJSON(user)),
		esClient.Index.WithContext(ctx),
		esClient.Index.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("error creating user: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error creating user: %s", res.String())
	}

	return nil
}

// GetUserByToken retrieves a user by their confirmation token
func GetUserByToken(token string) (*models.User, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				"confirmation_token": token,
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res, err := esClient.Search(
		esClient.Search.WithIndex(usersIndex),
		esClient.Search.WithBody(strings.NewReader(mustToJSON(query))),
		esClient.Search.WithContext(ctx),
	)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %v", err)
	}
	defer res.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	hits := result["hits"].(map[string]interface{})["hits"].([]interface{})
	if len(hits) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	source := hits[0].(map[string]interface{})["_source"]
	userData, err := json.Marshal(source)
	if err != nil {
		return nil, fmt.Errorf("error marshaling user data: %v", err)
	}

	var user models.User
	if err := json.Unmarshal(userData, &user); err != nil {
		return nil, fmt.Errorf("error unmarshaling user: %v", err)
	}

	return &user, nil
}

func GetUserByEmail(email string) (*models.User, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				"email": email,
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res, err := esClient.Search(
		esClient.Search.WithIndex(usersIndex),
		esClient.Search.WithBody(strings.NewReader(mustToJSON(query))),
		esClient.Search.WithContext(ctx),
	)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %v", err)
	}
	defer res.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	hits := result["hits"].(map[string]interface{})["hits"].([]interface{})
	if len(hits) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	source := hits[0].(map[string]interface{})["_source"]
	userData, err := json.Marshal(source)
	if err != nil {
		return nil, fmt.Errorf("error marshaling user data: %v", err)
	}

	var user models.User
	if err := json.Unmarshal(userData, &user); err != nil {
		return nil, fmt.Errorf("error unmarshaling user: %v", err)
	}

	return &user, nil
}

// ConfirmUser confirms a user's email and removes their confirmation token
func ConfirmUser(email string) error {
	script := map[string]interface{}{
		"script": map[string]interface{}{
			"source": "ctx._source.confirmed = true; ctx._source.confirmation_token = null",
		},
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				"email": email,
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res, err := esClient.UpdateByQuery(
		[]string{usersIndex},
		esClient.UpdateByQuery.WithBody(strings.NewReader(mustToJSON(script))),
		esClient.UpdateByQuery.WithContext(ctx),
		esClient.UpdateByQuery.WithRefresh(true),
	)
	if err != nil {
		return fmt.Errorf("error confirming user: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error confirming user: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return fmt.Errorf("error parsing response: %v", err)
	}

	// Check if any documents were updated
	updated, ok := result["updated"].(float64)
	if !ok || updated == 0 {
		return fmt.Errorf("user not found with email: %s", email)
	}

	return nil
}

// DeleteUnconfirmedUsers removes unconfirmed users older than the cutoff time
func DeleteUnconfirmedUsers(cutoff time.Time) error {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"term": map[string]interface{}{
							"confirmed": false,
						},
					},
					{
						"range": map[string]interface{}{
							"created_at": map[string]interface{}{
								"lt": cutoff.Format(time.RFC3339),
							},
						},
					},
				},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res, err := esClient.DeleteByQuery(
		[]string{usersIndex},
		strings.NewReader(mustToJSON(query)),
		esClient.DeleteByQuery.WithContext(ctx),
		esClient.DeleteByQuery.WithRefresh(true),
	)
	if err != nil {
		return fmt.Errorf("error deleting unconfirmed users: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error deleting unconfirmed users: %s", res.String())
	}

	return nil
}

// Helper function to convert interface to JSON string
func mustToJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("error marshaling to JSON: %v", err))
	}
	return string(b)
}

func Index(index string, body *strings.Reader) (*esapi.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res, err := esClient.Index(
		index,
		body,
		esClient.Index.WithContext(ctx),
		esClient.Index.WithRefresh("true"),
	)
	if err != nil {
		return nil, fmt.Errorf("error indexing document: %v", err)
	}

	return res, nil
}

func Count(index string) (*esapi.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res, err := esClient.Count(
		esClient.Count.WithIndex(index),
		esClient.Count.WithContext(ctx),
	)
	if err != nil {
		return nil, fmt.Errorf("error counting documents: %v", err)
	}

	return res, nil
}
