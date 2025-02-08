package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/sirupsen/logrus"
)

var jwtSecret []byte

// Claims represents the JWT claims structure
type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role,omitempty"`
	jwt.StandardClaims
}

/**
 * initialize the JWT secret key. This should be called once at the start of the application.
 */
func Initialize(secret string) {
	jwtSecret = []byte(secret)
}

/**
 * generate new JWT token for the given user ID and role. The token will expire in 24 hours.
 */
func GenerateToken(userID, role string) (string, error) {
	// set expiration to 24 hours
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		UserID: userID,
		Role:   role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

/**
 * parse the token string and validate it using the secret key. If the token is valid,
 * return the claims. If the token is invalid, return an error.
 */
func validateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

/**
 * checks for a valid JWT token in the Authorization header. If the token is valid,
 * extract the user ID and role from the token and add it to the request context.
 */
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			logrus.Error("No authorization header")
			returnGeneric401(w)
			return
		}

		// expected format: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			logrus.Error("Invalid authorization header format")
			returnGeneric401(w)
			return
		}

		claims, err := validateToken(parts[1])
		if err != nil {
			logrus.WithError(err).Error("Invalid token")
			returnGeneric401(w)
			return
		}

		// add claims to request context
		ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "role", claims.Role)

		// call the next handler with the enhanced context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func returnGeneric401(w http.ResponseWriter) {
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

/**
 * extract user ID from the request context. If the values are
 * not found, return an error.
 */
func GetUserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return "", errors.New("user ID not found in context")
	}
	return userID, nil
}

/**
 * extract role from the request context. If the values are
 * not found, return an error.
 */
func GetUserRoleFromContext(ctx context.Context) (string, error) {
	role, ok := ctx.Value("role").(string)
	if !ok {
		return "", errors.New("role not found in context")
	}
	return role, nil
}
