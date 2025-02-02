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

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	token, err := h.authService.AuthenticateUser(req.Email, req.Password)
	if err != nil {
		logrus.WithError(err).Error("Failed to authenticate user")
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	response := LoginResponse{Token: token}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Apply rate limiting
	if !h.authService.AllowRequest() {
		http.Error(w, "Too many requests", http.StatusTooManyRequests)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.authService.RegisterUser(req.Email, req.Password)
	if err != nil {
		switch err {
		case auth.ErrEmailExists:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case auth.ErrInvalidEmail:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case auth.ErrWeakPassword:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			logrus.WithError(err).Error("Failed to register user")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Registration successful. Please check your email to confirm your account.",
	})
}

func (h *AuthHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	if !h.authService.AllowRequest() {
		http.Error(w, "Too many requests", http.StatusTooManyRequests)
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, auth.ErrInvalidToken.Error(), http.StatusBadRequest)
		return
	}

	err := h.authService.ConfirmUser(token)
	if err != nil {
		switch err {
		case auth.ErrInvalidToken:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case auth.ErrTokenExpired:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			logrus.WithError(err).Error("Failed to confirm user")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Email confirmed successfully. You can now log in.",
	})
}
