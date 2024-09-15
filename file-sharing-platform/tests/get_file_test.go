package tests

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "file-sharing-platform/handlers"
    "file-sharing-platform/utils"
)

type contextKey string

const userContextKey = contextKey("user")

func TestGetFile(t *testing.T) {
    token, err := utils.GenerateJWT("test@example.com")
    if err != nil {
        t.Fatalf("Error generating token: %v", err)
    }
    
    req := httptest.NewRequest(http.MethodGet, "/file/1", nil)
    req.Header.Set("Authorization", "Bearer " + token)

    w := httptest.NewRecorder()

    ctx := context.WithValue(req.Context(), userContextKey, "test@example.com")

    handlers.GetFile(w, req.WithContext(ctx))

    res := w.Result()
    if res.StatusCode != http.StatusOK {
        t.Errorf("Expected status OK but got %v", res.Status)
    }
}
