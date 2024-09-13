package middleware

import (
    "context"
    "net/http"
    "strings"
    "file-sharing-platform/utils"
)

func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Authorization header is required", http.StatusUnauthorized)
            return
        }

        token := strings.Split(authHeader, " ")[1]
        if token == "" {
            http.Error(w, "Token is missing", http.StatusUnauthorized)
            return
        }

        claims, err := utils.ValidateJWT(token)
        if err != nil {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        ctx := context.WithValue(r.Context(), "user", claims.Email)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
