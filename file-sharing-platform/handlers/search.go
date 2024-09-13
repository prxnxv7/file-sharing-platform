package handlers

import (
    "context"
    "encoding/json"
    "net/http"
    "file-sharing-platform/config"
)

func SearchFiles(w http.ResponseWriter, r *http.Request) {
    db, err := config.ConnectDB()
    if err != nil {
        http.Error(w, "Error connecting to database", http.StatusInternalServerError)
        return
    }
    defer db.Close(context.Background())

    // Get search query parameters
    fileName := r.URL.Query().Get("file_name")
    fileType := r.URL.Query().Get("file_type")
    uploadDate := r.URL.Query().Get("upload_date")

    // Construct SQL query with conditions
    query := `SELECT id, file_name, file_size, s3_url FROM files WHERE 1=1`
    args := []interface{}{}
    
    if fileName != "" {
        query += ` AND file_name ILIKE '%' || $1 || '%'`
        args = append(args, fileName)
    }
    if fileType != "" {
        query += ` AND file_type = $2`
        args = append(args, fileType)
    }
    if uploadDate != "" {
        query += ` AND upload_date::DATE = $3`
        args = append(args, uploadDate)
    }

    rows, err := db.Query(context.Background(), query, args...)
    if err != nil {
        http.Error(w, "Error searching files", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    files := []map[string]interface{}{}
    for rows.Next() {
        var id int
        var fileName string
        var fileSize int64
        var s3URL string
        if err := rows.Scan(&id, &fileName, &fileSize, &s3URL); err != nil {
            http.Error(w, "Error retrieving file data", http.StatusInternalServerError)
            return
        }
        files = append(files, map[string]interface{}{
            "file_id":   id,
            "file_name": fileName,
            "file_size": fileSize,
            "s3_url":    s3URL,
        })
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(files)
}
