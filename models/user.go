package models

import "time"

type User struct {
	ID                string    `json:"id"`
	Email             string    `json:"email"`
	PasswordHash      string    `json:"password_hash"`
	Confirmed         bool      `json:"confirmed"`
	ConfirmationToken string    `json:"confirmation_token,omitempty"`
	TokenExpiry       time.Time `json:"token_expiry"`
	CreatedAt         time.Time `json:"created_at"`
}
