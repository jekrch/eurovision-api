package auth

import (
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

/**
 * AuthMiddleware is a middleware that checks for a valid JWT token in the
 * Authorization header of the request. If the token is valid, the request is
 * passed to the next handler. If the token is invalid, the middleware returns
 * a 401 Unauthorized response.
 */
func AuthMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			logrus.Error("No authorization header")
			returnGeneric401(w)
			return
		}

		// Expected format: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			logrus.Error("Invalid authorization header")
			returnGeneric401(w)
			return
		}

		token := parts[1]
		if !validateToken(token) {
			logrus.Error("Invalid token")
			returnGeneric401(w)
			return
		}

		// Add user info to request context if needed
		next.ServeHTTP(w, r)
	})
}

func returnGeneric401(w http.ResponseWriter) {
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

func validateToken(token string) bool {
	//TODO: Implement token validation
	return true
}
