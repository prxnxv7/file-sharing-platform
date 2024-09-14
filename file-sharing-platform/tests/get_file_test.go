package tests

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "file-sharing-platform/handlers"
    "file-sharing-platform/utils"
)

// Define a custom type for the context key
type contextKey string

const userContextKey = contextKey("user")

func TestGetFile(t *testing.T) {
    // Mock a valid JWT
    token, err := utils.GenerateJWT("test@example.com")
    if err != nil {
        t.Fatalf("Error generating token: %v", err)
    }
    
    req := httptest.NewRequest(http.MethodGet, "/file/1", nil)
    req.Header.Set("Authorization", "Bearer " + token)

    // Create a response recorder
    w := httptest.NewRecorder()

    // Create a test context with the user's email
    ctx := context.WithValue(req.Context(), userContextKey, "test@example.com")

    // Call the handler with the request and response recorder
    handlers.GetFile(w, req.WithContext(ctx))

    res := w.Result()
    if res.StatusCode != http.StatusOK {
        t.Errorf("Expected status OK but got %v", res.Status)
    }
}
