package tests

import (
    "bytes"
    // "io"
    "net/http"
    "net/http/httptest"
    "mime/multipart"
    "testing"

    "file-sharing-platform/handlers"
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

    req := httptest.NewRequest(http.MethodPost, "/upload/1", body)
    req.Header.Set("Content-Type", writer.FormDataContentType())
    w := httptest.NewRecorder()
    
    handlers.UploadFile(w, req)
    
    res := w.Result()
    if res.StatusCode != http.StatusOK {
        t.Errorf("Expected status OK but got %v", res.Status)
    }
}
