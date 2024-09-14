package tests

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "file-sharing-platform/handlers"
)

func TestLoginUser(t *testing.T) {
    reqBody := map[string]string{
        "email":    "user@example.com",
        "password": "password123",
    }
    body, _ := json.Marshal(reqBody)
    
    req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
    w := httptest.NewRecorder()
    
    handlers.LoginUser(w, req)
    
    res := w.Result()
    if res.StatusCode != http.StatusOK {
        t.Errorf("Expected status OK but got %v", res.Status)
    }
}
