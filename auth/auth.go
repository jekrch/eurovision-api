package auth

import (
	"errors"
	"eurovision-api/db"
	"fmt"
	"net/mail"
	"net/smtp"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

const (
	tokenExpiryHours = 24
	minPasswordLen   = 8
)

var (
	// Rate limiter: 3 attempts per minute per IP
	limiter = rate.NewLimiter(rate.Every(time.Minute/3), 1)

	ErrEmailExists  = errors.New("email already exists")
	ErrInvalidEmail = errors.New("invalid email format")
	ErrWeakPassword = errors.New("password too weak")
	ErrTokenExpired = errors.New("confirmation token expired")
	ErrInvalidToken = errors.New("invalid token")

	ErrInvalidCredentials     = errors.New("invalid email or password")
	ErrUnconfirmedEmail       = errors.New("email not confirmed")
	ErrUserNotFound           = errors.New("user not found")
	ErrRegistrationIncomplete = errors.New("registration not completed")
)

func generateConfirmationToken() (string, time.Time) {
	token := uuid.New().String()
	expiry := time.Now().Add(tokenExpiryHours * time.Hour)
	return token, expiry
}

func validateEmail(email string) error {
	_, err := mail.ParseAddress(email)
	return err
}

func validatePassword(password string) error {
	if len(password) < minPasswordLen {
		return ErrWeakPassword
	}
	return nil
}

func sendVerificationEmail(to, token string) error {

	baseURL := os.Getenv("APP_BASE_URL")
	verifyURL := fmt.Sprintf("%s/complete-registration?token=%s", baseURL, token)

	subject := "Complete Your Registration"
	body := fmt.Sprintf(`
		Hello Eurovision-Ranker user!
		
		Click the link below to verify your email and set your password:
		%s
		
		This link will expire in %d hours.
		
		If you didn't create this account, please ignore this email.
	`, verifyURL, tokenExpiryHours)

	return sendEmail(to, subject, body)
}

func sendPasswordResetEmail(to, token string) error {

	baseURL := os.Getenv("APP_BASE_URL")

	resetURL := fmt.Sprintf("%s/reset-password?token=%s", baseURL, token)

	subject := "Reset Your Password"
	body := fmt.Sprintf(`
		Hello Eurovision-Ranker user!
		
		Click the link below to reset your password:
		%s
		
		This link will expire in %d hours.
		
		If you didn't request this password reset, please ignore this email.
	`, resetURL, tokenExpiryHours)

	return sendEmail(to, subject, body)
}

func sendEmail(to, subject, body string) error {

	from := os.Getenv("EMAIL_USER")
	password := os.Getenv("EMAIL_PASSWORD")
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")

	msg := fmt.Sprintf("To: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		to, subject, body)

	auth := smtp.PlainAuth("", from, password, host)

	return smtp.SendMail(host+":"+port, auth, from, []string{to}, []byte(msg))
}

// Cleanup job to remove unconfirmed users
func cleanupUnconfirmedUsers() {
	cutoff := time.Now().Add(-48 * time.Hour)
	if err := db.DeleteUnconfirmedUsers(cutoff); err != nil {
		logrus.Error("Failed to cleanup unconfirmed users", "error", err)
	}
}
