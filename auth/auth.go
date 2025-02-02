package auth

import (
	"errors"
	"eurovision-api/db"
	"eurovision-api/models"
	"fmt"
	"net/http"
	"net/mail"
	"net/smtp"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
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

	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUnconfirmedEmail   = errors.New("email not confirmed")
	ErrUserNotFound       = errors.New("user not found")
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
	// Add more password strength requirements here
	return nil
}

func sendConfirmationEmail(to, token string) error {
	from := os.Getenv("EMAIL_USER")
	password := os.Getenv("EMAIL_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")

	baseURL := os.Getenv("APP_BASE_URL")
	confirmURL := fmt.Sprintf("%s/confirm?token=%s", baseURL, token)

	subject := "Confirm Your Email"
	body := fmt.Sprintf(`
		Hello!
		
		Please confirm your email by clicking the link below:
		%s
		
		This link will expire in %d hours.
		
		If you didn't create this account, please ignore this email.
	`, confirmURL, tokenExpiryHours)

	msg := fmt.Sprintf("To: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		to, subject, body)

	auth := smtp.PlainAuth("", from, password, smtpHost)
	return smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, []byte(msg))
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	// Apply rate limiting
	if !limiter.Allow() {
		http.Error(w, "Too many requests", http.StatusTooManyRequests)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	if err := validateEmail(email); err != nil {
		http.Error(w, ErrInvalidEmail.Error(), http.StatusBadRequest)
		return
	}

	if err := validatePassword(password); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logrus.Error("Failed to hash password", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Check if email exists using proper index
	exists, err := db.EmailExists(email)
	if err != nil {
		logrus.Error("Failed to check email existence", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, ErrEmailExists.Error(), http.StatusBadRequest)
		return
	}

	token, expiry := generateConfirmationToken()

	user := models.User{
		Email:             email,
		PasswordHash:      string(hashedPassword),
		Confirmed:         false,
		ConfirmationToken: token,
		TokenExpiry:       expiry,
		CreatedAt:         time.Now(),
	}

	if err := db.CreateUser(&user); err != nil {
		logrus.Error("Failed to create user", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := sendConfirmationEmail(email, token); err != nil {
		logrus.Error("Failed to send confirmation email", "error", err)
		// Consider rolling back user creation here
		http.Error(w, "Failed to send confirmation email", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Registration successful. Please check your email to confirm your account.")
}

func confirmHandler(w http.ResponseWriter, r *http.Request) {
	if !limiter.Allow() {
		http.Error(w, "Too many requests", http.StatusTooManyRequests)
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, ErrInvalidToken.Error(), http.StatusBadRequest)
		return
	}

	user, err := db.GetUserByToken(token)
	if err != nil {
		logrus.Error("Failed to get user by token", "error", err)
		http.Error(w, ErrInvalidToken.Error(), http.StatusBadRequest)
		return
	}

	if user.TokenExpiry.Before(time.Now()) {
		http.Error(w, ErrTokenExpired.Error(), http.StatusBadRequest)
		return
	}

	// Atomic update
	if err := db.ConfirmUser(user.Email); err != nil {
		logrus.Error("Failed to confirm user", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Email confirmed successfully. You can now log in.")
}

// Cleanup job to remove unconfirmed users
func cleanupUnconfirmedUsers() {
	cutoff := time.Now().Add(-48 * time.Hour)
	if err := db.DeleteUnconfirmedUsers(cutoff); err != nil {
		logrus.Error("Failed to cleanup unconfirmed users", "error", err)
	}
}
