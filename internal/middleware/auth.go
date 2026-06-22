package middleware

import (
	"context"
	"habit-tracker/internal/auth"
	"net/http"
)

// Ключ для контекста
type contextKey string

const userIDKey contextKey = "userID"
const roleKey contextKey = "role"

// AuthMiddleware проверяет JWT токен из заголовка Authorization
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		token := authHeader
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		}

		userID, role, err := auth.ValidateToken(token)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		ctx = context.WithValue(ctx, roleKey, role)

		next(w, r.WithContext(ctx))
	}
}

// GetUserIDFromContext извлекает ID пользователя из контекста
func GetUserIDFromContext(r *http.Request) (int, bool) {
	userID, ok := r.Context().Value(userIDKey).(int)
	return userID, ok
}

// GetRoleFromContext извлекает роль пользователя из контекста
func GetRoleFromContext(r *http.Request) (string, bool) {
	role, ok := r.Context().Value(roleKey).(string)
	return role, ok
}

// RequireRole разрешает доступ к маршруту только пользователям с одной из указанных ролей.
func RequireRole(allowedRoles ...string) func(http.HandlerFunc) http.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedRoles))
	for _, role := range allowedRoles {
		allowed[role] = struct{}{}
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			role, ok := GetRoleFromContext(r)
			if !ok {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			if _, ok := allowed[role]; !ok {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next(w, r)
		}
	}
}
