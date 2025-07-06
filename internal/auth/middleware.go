package auth

import (
	"context"
	"net/http"
	"strings"
	"messenger/internal/utils"
)

type contextKey string
const UserIDKey contextKey = "user_id"

// AuthMiddleware проверяет JWT-токен и добавляет user_id в контекст
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		userID, err := utils.ValidateJWT(tokenString)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
