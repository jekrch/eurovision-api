package auth

import (
	"eurovision-api/db"
	"eurovision-api/models"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/time/rate"
)

type Service struct {
	limiter   *rate.Limiter
	jwtSecret []byte
}

func NewService(jwtSecret string) *Service {
	return &Service{
		limiter:   rate.NewLimiter(rate.Every(time.Minute/10), 3),
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *Service) AllowRequest() bool {
	return s.limiter.Allow()
}

func (s *Service) RegisterUser(email, password string) error {
	if err := validateEmail(email); err != nil {
		return ErrInvalidEmail
	}

	if err := validatePassword(password); err != nil {
		return err
	}

	exists, err := db.EmailExists(email)
	if err != nil {
		return err
	}
	if exists {
		return ErrEmailExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	token, expiry := generateConfirmationToken()

	user := models.User{
		ID:                uuid.New().String(),
		Email:             email,
		PasswordHash:      string(hashedPassword),
		Confirmed:         false,
		ConfirmationToken: token,
		TokenExpiry:       expiry,
		CreatedAt:         time.Now(),
	}

	if err := db.CreateUser(&user); err != nil {
		return err
	}

	return sendConfirmationEmail(email, token)
}

func (s *Service) ConfirmUser(token string) error {
	user, err := db.GetUserByToken(token)
	if err != nil {
		return ErrInvalidToken
	}

	if user.TokenExpiry.Before(time.Now()) {
		return ErrTokenExpired
	}

	return db.ConfirmUser(user.Email)
}

func (s *Service) AuthenticateUser(email, password string) (string, error) {
	user, err := db.GetUserByEmail(email)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	if !user.Confirmed {
		return "", ErrUnconfirmedEmail
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":   user.Email,
		"user_id": user.ID,
		"role":    "user",
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	return token.SignedString(s.jwtSecret)
}
