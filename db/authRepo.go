package db

import (
	"context"
	"encoding/json"
	"eurovision-api/models"
	"fmt"
	"time"

	"github.com/olivere/elastic/v7"
)

/**
 * creates the users index with proper mappings if it doesn't exist.
 */
func createUsersIndex() error {

	mapping := `{
		"mappings": {
			"properties": {
				"id": {
					"type": "keyword"
				},
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

	return createIndex(usersIndex, mapping)
}

/**
 * updates the user's password and confirms their email
 */
func CompleteRegistration(email, passwordHash string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	script := elastic.NewScript(`
		ctx._source.password_hash = params.password_hash;
		ctx._source.confirmed = true;
		ctx._source.confirmation_token = null;
	`).Param("password_hash", passwordHash)

	query := elastic.NewTermQuery("email", email)

	result, err := esClient.UpdateByQuery(usersIndex).
		Query(query).
		Script(script).
		Refresh("true").
		Do(ctx)

	if err != nil {
		return fmt.Errorf("error completing registration: %v", err)
	}

	if result.Updated == 0 {
		return fmt.Errorf("user not found with email: %s", email)
	}

	return nil
}

/**
 * updates the user's reset token and expiry
 */
func SetResetToken(email, token string, expiry time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	script := elastic.NewScript(`
		ctx._source.confirmation_token = params.token;
		ctx._source.token_expiry = params.expiry;
	`).Params(map[string]interface{}{
		"token":  token,
		"expiry": expiry,
	})

	query := elastic.NewTermQuery("email", email)

	result, err := esClient.UpdateByQuery(usersIndex).
		Query(query).
		Script(script).
		Refresh("true").
		Do(ctx)

	if err != nil {
		return fmt.Errorf("error setting reset token: %v", err)
	}

	if result.Updated == 0 {
		return fmt.Errorf("user not found with email: %s", email)
	}

	return nil
}

/**
 * updates the user's password and removes the reset token
 */
func UpdatePassword(email, passwordHash string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	script := elastic.NewScript(`
		ctx._source.password_hash = params.password_hash;
		ctx._source.confirmation_token = null;
	`).Param("password_hash", passwordHash)

	query := elastic.NewTermQuery("email", email)

	result, err := esClient.UpdateByQuery(usersIndex).
		Query(query).
		Script(script).
		Refresh("true").
		Do(ctx)

	if err != nil {
		return fmt.Errorf("error updating password: %v", err)
	}

	if result.Updated == 0 {
		return fmt.Errorf("user not found with email: %s", email)
	}

	return nil
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
