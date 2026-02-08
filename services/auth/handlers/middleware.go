package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserContextKey contextKey = "user"

type UserClaims struct {
	UserID string
	Email  string
	Role   string
}

func (h *AuthHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing authorization header"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid authorization format"})
			return
		}

		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return h.JWTSecret, nil
		})
		if err != nil || !token.Valid {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token claims"})
			return
		}

		userClaims := UserClaims{
			UserID: claims["user_id"].(string),
			Email:  claims["email"].(string),
			Role:   claims["role"].(string),
		}

		ctx := context.WithValue(r.Context(), UserContextKey, userClaims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *AuthHandler) Validate(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(UserContextKey).(UserClaims)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
		return
	}
	writeJSON(w, http.StatusOK, user)
}
