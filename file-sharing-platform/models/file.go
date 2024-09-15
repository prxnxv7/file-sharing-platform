package models

import (
    "time"
)

type File struct {
    ID        int       `json:"id" db:"id"`
    UserID    int       `json:"user_id" db:"user_id"`
    FileName  string    `json:"file_name" db:"file_name"`
    FileSize  int64     `json:"file_size" db:"file_size"`
    S3URL     string    `json:"s3_url" db:"s3_url"`
    UploadDate time.Time `json:"upload_date" db:"upload_date"`
}
