package handlers

import (
	"context"
	"encoding/json"
	"file-sharing-platform/config"
	"net/http"
	"strconv"
	"strings"
)

func SearchFiles(w http.ResponseWriter, r *http.Request) {
    db, err := config.ConnectDB()
    if err != nil {
        http.Error(w, "Error connecting to database", http.StatusInternalServerError)
        return
    }
    defer db.Close(context.Background())

    fileName := r.URL.Query().Get("file_name")
    fileType := r.URL.Query().Get("file_type")
    uploadDate := r.URL.Query().Get("upload_date")

    query := `SELECT id, file_name, file_size, local_path FROM files WHERE 1=1`
    args := []interface{}{}
    var conditions []string

    if fileName != "" {
        conditions = append(conditions, `file_name ILIKE '%' || $`+strconv.Itoa(len(args)+1)+` || '%'`)
        args = append(args, fileName)
    }
    if fileType != "" {
        conditions = append(conditions, `file_type = $`+strconv.Itoa(len(args)+1))
        args = append(args, fileType)
    }
    if uploadDate != "" {
        conditions = append(conditions, `upload_date::DATE = $`+strconv.Itoa(len(args)+1))
        args = append(args, uploadDate)
    }

    if len(conditions) > 0 {
        query += " AND " + strings.Join(conditions, " AND ")
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
        var localPath string
        if err := rows.Scan(&id, &fileName, &fileSize, &localPath); err != nil {
            http.Error(w, "Error retrieving file data", http.StatusInternalServerError)
            return
        }
        files = append(files, map[string]interface{}{
            "file_id":   id,
            "file_name": fileName,
            "file_size": fileSize,
            "local_path": localPath,
        })
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(files)
}
