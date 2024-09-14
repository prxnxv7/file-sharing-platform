package tests

import (
	"bytes"
	"context"
	// "io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"file-sharing-platform/handlers"
	"file-sharing-platform/utils"
)

func TestUploadFile(t *testing.T) {
    body := new(bytes.Buffer)
    writer := multipart.NewWriter(body)
    file, err := writer.CreateFormFile("file", "testfile.txt")
    if err != nil {
        t.Fatalf("Error creating form file: %v", err)
    }
    file.Write([]byte("This is a test file"))
    writer.Close()

    // Mock a valid JWT
    token, err := utils.GenerateJWT("test@example.com")
    if err != nil {
        t.Fatalf("Error generating token: %v", err)
    }

    req := httptest.NewRequest(http.MethodPost, "/upload/1", body)
    req.Header.Set("Content-Type", writer.FormDataContentType())
    req.Header.Set("Authorization", "Bearer " + token)
    w := httptest.NewRecorder()
    
    // Mock context
    ctx := context.WithValue(req.Context(), userContextKey, "test@example.com")

    handlers.UploadFile(w, req.WithContext(ctx))
    
    res := w.Result()
    if res.StatusCode != http.StatusOK {
        t.Errorf("Expected status OK but got %v", res.Status)
    }
}
