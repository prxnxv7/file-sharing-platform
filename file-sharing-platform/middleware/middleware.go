package middleware

import (
	"context"
	"file-sharing-platform/services"
	"file-sharing-platform/utils"
	"net/http"
	"strings"
	"time"
)

// Define a custom type for the context key to avoid collisions
type contextKey string

const (
    userContextKey  contextKey = "user" 
	requestLimit   = 100              // Number of allowed requests per window
	windowDuration = time.Minute       // Time window for rate limiting
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

// RateLimiterMiddleware limits the number of requests per user based on the user’s email in JWT claims
func RateLimiterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Retrieve user email from the context set by the AuthMiddleware
		userEmail, ok := r.Context().Value("user").(string)
		if !ok || userEmail == "" {
			http.Error(w, "User not authenticated", http.StatusUnauthorized)
			return
		}

		// Apply rate limiting logic using Redis
		allowed, err := services.RateLimit(userEmail, requestLimit, windowDuration)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// If rate limit exceeded, return 429 Too Many Requests
		if !allowed {
			http.Error(w, "Rate limit exceeded, try again later", http.StatusTooManyRequests)
			return
		}

		// Proceed to the next handler
		next.ServeHTTP(w, r)
	})
}