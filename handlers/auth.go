package handlers

import (
	"encoding/json"
	"eurovision-api/auth"
	"net/http"

	"github.com/sirupsen/logrus"
)

type AuthHandler struct {
	authService *auth.Service
}

func NewAuthHandler(authService *auth.Service) *AuthHandler {
	if authService == nil {
		panic("auth service cannot be nil")
	}
	return &AuthHandler{
		authService: authService,
	}
}

// Request/Response structs
type InitiateRegistrationRequest struct {
	Email string `json:"email"`
}

type CompleteRegistrationRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type InitiatePasswordResetRequest struct {
	Email string `json:"email"`
}

type CompletePasswordResetRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// InitiateRegistration handles the first step of registration (email only)
func (h *AuthHandler) InitiateRegistration(w http.ResponseWriter, r *http.Request) {
	if !h.authService.AllowRequest() {
		http.Error(w, "Too many requests", http.StatusTooManyRequests)
		return
	}

	var req InitiateRegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.authService.InitiateRegistration(req.Email)
	if err != nil {
		switch err {
		case auth.ErrEmailExists:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case auth.ErrInvalidEmail:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			logrus.WithError(err).Error("Failed to initiate registration")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Please check your email to complete registration.",
	})
}

// CompleteRegistration handles setting the password after email verification
func (h *AuthHandler) CompleteRegistration(w http.ResponseWriter, r *http.Request) {
	if !h.authService.AllowRequest() {
		http.Error(w, "Too many requests", http.StatusTooManyRequests)
		return
	}

	var req CompleteRegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.authService.CompleteRegistration(req.Token, req.Password)
	if err != nil {
		switch err {
		case auth.ErrInvalidToken:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case auth.ErrTokenExpired:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case auth.ErrWeakPassword:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			logrus.WithError(err).Error("Failed to complete registration")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Registration completed successfully. You can now log in.",
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if !h.authService.AllowRequest() {
		http.Error(w, "Too many requests", http.StatusTooManyRequests)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	token, err := h.authService.AuthenticateUser(req.Email, req.Password)
	if err != nil {
		switch err {
		case auth.ErrInvalidCredentials:
			http.Error(w, err.Error(), http.StatusUnauthorized)
		case auth.ErrUnconfirmedEmail:
			http.Error(w, err.Error(), http.StatusForbidden)
		case auth.ErrRegistrationIncomplete:
			http.Error(w, err.Error(), http.StatusForbidden)
		default:
			logrus.WithError(err).Error("Failed to authenticate user")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	response := LoginResponse{Token: token}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) InitiatePasswordReset(w http.ResponseWriter, r *http.Request) {
	if !h.authService.AllowRequest() {
		http.Error(w, "Too many requests", http.StatusTooManyRequests)
		return
	}

	var req InitiatePasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.authService.InitiatePasswordReset(req.Email)
	if err != nil {
		// Don't reveal if email exists or not
		logrus.WithError(err).Error("Failed to initiate password reset")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "If your email exists in our system, you will receive password reset instructions.",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "If your email exists in our system, you will receive password reset instructions.",
	})
}

func (h *AuthHandler) CompletePasswordReset(w http.ResponseWriter, r *http.Request) {
	if !h.authService.AllowRequest() {
		http.Error(w, "Too many requests", http.StatusTooManyRequests)
		return
	}

	var req CompletePasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.authService.CompletePasswordReset(req.Token, req.NewPassword)
	if err != nil {
		switch err {
		case auth.ErrInvalidToken:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case auth.ErrTokenExpired:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case auth.ErrWeakPassword:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			logrus.WithError(err).Error("Failed to complete password reset")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Password has been reset successfully. You can now log in with your new password.",
	})
}
