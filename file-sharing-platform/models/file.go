package models

// File represents a file uploaded by a user
type File struct {
    ID        int    `json:"id"`
    UserID    int    `json:"user_id"`
    FileName  string `json:"file_name"`
    FileSize  int64  `json:"file_size"`
    UploadDate string `json:"upload_date"`
    S3URL     string `json:"local_path"`
}
