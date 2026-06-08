package auth

import (
    "context"
    "net/http"
    "strings"
)

type contextKey string

const UserContextKey contextKey = "userID"

// AuthMiddleware проверяет наличие и валидность JWT в заголовке Authorization
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Missing token", http.StatusUnauthorized)
            return
        }
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            http.Error(w, "Invalid token format", http.StatusUnauthorized)
            return
        }
        claims, err := ValidateToken(parts[1])
        if err != nil {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }
        ctx := context.WithValue(r.Context(), UserContextKey, claims.UserID)
        next(w, r.WithContext(ctx))
    }
}

// GetUserID возвращает ID пользователя из контекста (после обработки AuthMiddleware)
func GetUserID(r *http.Request) string {
    if v := r.Context().Value(UserContextKey); v != nil {
        return v.(string)
    }
    return ""
}