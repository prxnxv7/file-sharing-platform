package tests

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "file-sharing-platform/handlers"
)

func TestRegisterUser(t *testing.T) {
    reqBody := map[string]string{
        "email":    "test@example.com",
        "password": "password123",
    }
    body, _ := json.Marshal(reqBody)
    
    req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
    w := httptest.NewRecorder()
    
    handlers.RegisterUser(w, req)
    
    res := w.Result()
    if res.StatusCode != http.StatusOK {
        t.Errorf("Expected status OK but got %v", res.Status)
    }
}
