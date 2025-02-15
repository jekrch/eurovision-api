package auth

import (
	"eurovision-api/db"
	"eurovision-api/models"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
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

/**
 * validates the email and checks if it already exists in the database. if not, a
 * confirmation token is generated and saved in the database. The token is
 * then sent to the user in an email.
 */
func (s *Service) InitiateRegistration(email string) error {
	if err := validateEmail(email); err != nil {
		return ErrInvalidEmail
	}

	exists, err := db.EmailExists(email)
	if err != nil {
		return err
	}
	if exists {
		return ErrEmailExists
	}

	token, expiry := generateConfirmationToken()

	user := models.User{
		ID:                uuid.New().String(),
		Email:             email,
		PasswordHash:      "", // Password will be set later
		Confirmed:         false,
		ConfirmationToken: token,
		TokenExpiry:       expiry,
		CreatedAt:         time.Now(),
	}

	if err := db.CreateUser(&user); err != nil {
		return err
	}

	return sendVerificationEmail(email, token)
}

/**
 * validates the token and sets the password for the user
 */
func (s *Service) CompleteRegistration(token, password string) error {
	if err := validatePassword(password); err != nil {
		return err
	}

	user, err := db.GetUserByToken(token)
	if err != nil {
		return ErrInvalidToken
	}

	if user.TokenExpiry.Before(time.Now()) {
		return ErrTokenExpired
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return db.CompleteRegistration(user.Email, string(hashedPassword))
}

/**
 * generates a new token and sends a password reset email
 */
func (s *Service) InitiatePasswordReset(email string) error {
	if err := validateEmail(email); err != nil {
		return ErrInvalidEmail
	}

	user, err := db.GetUserByEmail(email)
	if err != nil {
		logrus.Infof("Password reset requested for non-existent email: %s", email)
		return nil // Don't reveal if email exists
	}

	token, expiry := generateConfirmationToken()

	if err := db.SetResetToken(user.Email, token, expiry); err != nil {
		return err
	}

	return sendPasswordResetEmail(email, token)
}

/**
 * validates the reset token and sets the new password
 */
func (s *Service) CompletePasswordReset(token, newPassword string) error {
	if err := validatePassword(newPassword); err != nil {
		return err
	}

	user, err := db.GetUserByToken(token)
	if err != nil {
		return ErrInvalidToken
	}

	if user.TokenExpiry.Before(time.Now()) {
		return ErrTokenExpired
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return db.UpdatePassword(user.Email, string(hashedPassword))
}

func (s *Service) AuthenticateUser(email, password string) (string, error) {
	user, err := db.GetUserByEmail(email)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	if !user.Confirmed {
		return "", ErrUnconfirmedEmail
	}

	if user.PasswordHash == "" {
		return "", ErrRegistrationIncomplete
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":   user.Email,
		"user_id": user.ID,
		"role":    "user",
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	return token.SignedString(s.jwtSecret)
}
