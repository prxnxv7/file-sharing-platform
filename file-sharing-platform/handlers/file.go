package handlers

import (
    "context"
    "database/sql"
    "encoding/json"
    "net/http"
    "file-sharing-platform/config"
    "file-sharing-platform/services"
    "github.com/gorilla/mux"
    "strconv"
)

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

    // Upload file to S3 (or local)
    fileURL, err := services.UploadToS3(file, fileHeader, userID)
    if err != nil {
        http.Error(w, "Error uploading file", http.StatusInternalServerError)
        return
    }

    // Connect to the database
    db, err := config.ConnectDB()
    if err != nil {
        http.Error(w, "Error connecting to database", http.StatusInternalServerError)
        return
    }
    defer db.Close(context.Background())

    query := `INSERT INTO files (user_id, file_name, file_size, s3_url) VALUES ($1, $2, $3, $4)`
    _, err = db.Exec(context.Background(), query, userID, fileHeader.Filename, fileHeader.Size, fileURL)
    if err != nil {
        http.Error(w, "Error saving file metadata", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("File uploaded successfully: " + fileURL))
}

func GetFile(w http.ResponseWriter, r *http.Request) {
    fileID, err := strconv.Atoi(mux.Vars(r)["file_id"])
    if err != nil {
        http.Error(w, "Invalid file ID", http.StatusBadRequest)
        return
    }

    db, err := config.ConnectDB()
    if err != nil {
        http.Error(w, "Error connecting to database", http.StatusInternalServerError)
        return
    }
    defer db.Close(context.Background())

    query := `SELECT file_name, file_size, s3_url FROM files WHERE id = $1`
    row := db.QueryRow(context.Background(), query, fileID)

    var fileName string
    var fileSize int64
    var s3URL string
    err = row.Scan(&fileName, &fileSize, &s3URL)
    if err != nil {
        if err == sql.ErrNoRows {
            http.Error(w, "File not found", http.StatusNotFound)
        } else {
            http.Error(w, "Error retrieving file", http.StatusInternalServerError)
        }
        return
    }

    response := map[string]interface{}{
        "file_name": fileName,
        "file_size": fileSize,
        "s3_url":    s3URL,
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
