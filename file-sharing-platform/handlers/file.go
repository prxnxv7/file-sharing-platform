package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"file-sharing-platform/config"
	"file-sharing-platform/services"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// UploadFile handles file uploads and saves metadata
func UploadFile(w http.ResponseWriter, r *http.Request) {
    userID, err := strconv.Atoi(mux.Vars(r)["user_id"])
    if err != nil {
        http.Error(w, "Invalid user ID", http.StatusBadRequest)
        return
    }

    file, fileHeader, err := r.FormFile("file")
    if err != nil {
        http.Error(w, "Error reading file", http.StatusBadRequest)
        return
    }
    defer file.Close()

    // Create a channel to receive the result of file processing
    resultChan := make(chan error)

    // Start a goroutine for processing the file upload
    go func() {
        // Define the local storage path
        localDir := "./uploads" // Adjust this path as needed
        if err := os.MkdirAll(localDir, os.ModePerm); err != nil {
            resultChan <- err
            return
        }

        // Save the file to local storage
        localPath := filepath.Join(localDir, fileHeader.Filename)
        outFile, err := os.Create(localPath)
        if err != nil {
            resultChan <- err
            return
        }
        defer outFile.Close()

        _, err = io.Copy(outFile, file)
        if err != nil {
            resultChan <- err
            return
        }

        // Connect to the database
        db, err := config.ConnectDB()
        if err != nil {
            resultChan <- err
            return
        }
        defer db.Close(context.Background())

        query := `INSERT INTO files (user_id, file_name, file_size, local_path) VALUES ($1, $2, $3, $4)`
        _, err = db.Exec(context.Background(), query, userID, fileHeader.Filename, fileHeader.Size, localPath)
        if err != nil {
            resultChan <- err
            return
        }

        // Cache the file metadata in Redis
        fileMetadata := map[string]interface{}{
            "user_id":    userID,
            "file_name":  fileHeader.Filename,
            "file_size":  fileHeader.Size,
            "local_path": localPath,
        }
        metadataJSON, _ := json.Marshal(fileMetadata)
        cacheKey := "file_metadata:" + fileHeader.Filename
        cacheErr := services.CacheFileMetadata(cacheKey, string(metadataJSON), 5*time.Minute)
        if cacheErr != nil {
            resultChan <- cacheErr
            return
        }

        resultChan <- nil
        // Notify the WebSocket handler that the upload is complete
        uploadCompleteChan <- true
    }()

    // Respond to the client while file processing happens in the background
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("File upload started"))

    // Handle any errors from the goroutine
    if err := <-resultChan; err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        uploadCompleteChan <- false // Notify WebSocket on error
    }
}

// GetFile handles file retrieval
func GetFile(w http.ResponseWriter, r *http.Request) {
    fileID, err := strconv.Atoi(mux.Vars(r)["file_id"])
    if err != nil {
        http.Error(w, "Invalid file ID", http.StatusBadRequest)
        return
    }

    // Define Redis cache key
    cacheKey := "file_metadata:" + strconv.Itoa(fileID)

    // Try to get the metadata from Redis cache
    cachedMetadata, cacheErr := services.GetCachedFileMetadata(cacheKey)
    if cacheErr == nil && cachedMetadata != "" {
        // Cache hit - use cached metadata
        var metadata map[string]interface{}
        json.Unmarshal([]byte(cachedMetadata), &metadata)
        localPath := metadata["local_path"].(string)

        filePath := filepath.Join("uploads", localPath)
        if _, err := os.Stat(filePath); os.IsNotExist(err) {
            http.Error(w, "File does not exist", http.StatusNotFound)
            return
        }

        // Serve the file from local storage
        http.ServeFile(w, r, filePath)
        return
    }

    db, err := config.ConnectDB()
    if err != nil {
        http.Error(w, "Error connecting to database", http.StatusInternalServerError)
        return
    }
    defer db.Close(context.Background())

    query := `SELECT file_name, file_size, local_path FROM files WHERE id = $1`
    row := db.QueryRow(context.Background(), query, fileID)

    var fileName string
    var fileSize int64
    var localPath string
    err = row.Scan(&fileName, &fileSize, &localPath)
    if err != nil {
        if err == sql.ErrNoRows {
            http.Error(w, "File not found", http.StatusNotFound)
        } else {
            http.Error(w, "Error retrieving file", http.StatusInternalServerError)
        }
        return
    }

    // Cache the retrieved metadata in Redis
    fileMetadata := map[string]interface{}{
        "file_name":  fileName,
        "file_size":  fileSize,
        "local_path": localPath,
    }
    metadataJSON, _ := json.Marshal(fileMetadata)
    cacheErr = services.CacheFileMetadata(cacheKey, string(metadataJSON), 5*time.Minute)
    if cacheErr != nil {
        http.Error(w, "Error caching file metadata", http.StatusInternalServerError)
        return
    }

    filePath := filepath.Join("uploads", localPath)
    if _, err := os.Stat(filePath); os.IsNotExist(err) {
        http.Error(w, "File does not exist", http.StatusNotFound)
        return
    }

    http.ServeFile(w, r, filePath)
}
