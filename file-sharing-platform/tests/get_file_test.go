package tests

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "file-sharing-platform/handlers"
)

func TestGetFile(t *testing.T) {
    req := httptest.NewRequest(http.MethodGet, "/file/1", nil)
    w := httptest.NewRecorder()
    
    handlers.GetFile(w, req)
    
    res := w.Result()
    if res.StatusCode != http.StatusOK {
        t.Errorf("Expected status OK but got %v", res.Status)
    }
}
