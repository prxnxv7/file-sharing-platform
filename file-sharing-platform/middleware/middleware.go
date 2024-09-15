package middleware

import (
	"context"
	"file-sharing-platform/services"
	"file-sharing-platform/utils"
	"net/http"
	"strings"
	"time"
)

type contextKey string

const (
    userContextKey  contextKey = "user" 
	requestLimit   = 100             
	windowDuration = time.Minute     
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

        ctx := context.WithValue(r.Context(), userContextKey, claims.Email)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func RateLimiterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userEmail, ok := r.Context().Value("user").(string)
		if !ok || userEmail == "" {
			http.Error(w, "User not authenticated", http.StatusUnauthorized)
			return
		}
		allowed, err := services.RateLimit(userEmail, requestLimit, windowDuration)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if !allowed {
			http.Error(w, "Rate limit exceeded, try again later", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}