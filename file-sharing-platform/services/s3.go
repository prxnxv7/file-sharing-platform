// package services

// import (
// 	"bytes"
// 	"context"
// 	"fmt"
// 	"log"
// 	"mime/multipart"
// 	"net/http"
// 	"os"
// 	"strings"
// 	"time"

// 	"database/sql"

// 	"github.com/aws/aws-sdk-go/aws"
// 	"github.com/aws/aws-sdk-go/aws/session"
// 	"github.com/aws/aws-sdk-go/service/s3"
// 	_ "github.com/lib/pq"
// )

// var (
// 	S3Region = os.Getenv("AWS_REGION")
// 	S3Bucket = os.Getenv("AWS_BUCKET")
// 	DBConn   *sql.DB
// )

// var svc *s3.S3

// func init() {
// 	var err error
// 	DBConn, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
// 	if err != nil {
// 		log.Fatalf("Failed to connect to database: %v", err)
// 	}
// }

// func UploadToS3(file multipart.File, fileHeader *multipart.FileHeader, userID int) (string, error) {
// 	sess, err := session.NewSession(&aws.Config{
// 		Region: aws.String(S3Region),
// 	})
// 	if err != nil {
// 		return "", err
// 	}

// 	svc := s3.New(sess)

// 	buffer := make([]byte, fileHeader.Size)
// 	file.Read(buffer)

// 	s3Path := fmt.Sprintf("uploads/user-%d/%s", userID, fileHeader.Filename)

// 	_, err = svc.PutObject(&s3.PutObjectInput{
// 		Bucket:             aws.String(S3Bucket),
// 		Key:                aws.String(s3Path),
// 		ACL:                aws.String("public-read"),
// 		Body:               bytes.NewReader(buffer),
// 		ContentLength:      aws.Int64(fileHeader.Size),
// 		ContentType:        aws.String(http.DetectContentType(buffer)),
// 		ContentDisposition: aws.String("attachment"),
// 	})

// 	if err != nil {
// 		return "", err
// 	}

// 	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", S3Bucket, S3Region, s3Path), nil
// }

// func CleanUpExpiredFiles() {
// 	ctx := context.Background()

// 	// Example: Get files older than 30 days
// 	threshold := time.Now().Add(-30 * 24 * time.Hour)
// 	rows, err := DBConn.QueryContext(ctx, "SELECT id, s3_url FROM files WHERE upload_date < $1", threshold)
// 	if err != nil {
// 		log.Printf("Failed to query files for cleanup: %v", err)
// 		return
// 	}
// 	defer rows.Close()

// 	for rows.Next() {
// 		var id int
// 		var s3URL string
// 		if err := rows.Scan(&id, &s3URL); err != nil {
// 			log.Printf("Failed to scan file row: %v", err)
// 			continue
// 		}

// 		// Delete from S3
// 		s3Path := s3URL[strings.LastIndex(s3URL, "/")+1:]
// 		_, err = svc.DeleteObject(&s3.DeleteObjectInput{
// 			Bucket: aws.String(S3Bucket),
// 			Key:    aws.String(s3Path),
// 		})
// 		if err != nil {
// 			log.Printf("Failed to delete S3 object: %v", err)
// 		}

// 		// Delete from PostgreSQL
// 		_, err = DBConn.ExecContext(ctx, "DELETE FROM files WHERE id = $1", id)
// 		if err != nil {
// 			log.Printf("Failed to delete file from database: %v", err)
// 		}
// 	}
// }

package services

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v4"
)

var (
    CheckInterval = 24 * time.Hour
    UploadDir     = "uploads"
)

// UploadFile handles uploading a file to local storage and returns the file path
func UploadFile(file multipart.File, fileHeader *multipart.FileHeader, userID int) (string, error) {
	// Ensure the directory exists
	err := os.MkdirAll(UploadDir, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("failed to create upload directory: %v", err)
	}

	// Create the file path with user ID and file name
	filePath := filepath.Join(UploadDir, fmt.Sprintf("user-%d-%s", userID, fileHeader.Filename))

	// Create the file on the local system
	outFile, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer outFile.Close()

	// Copy the file content to the created file
	if _, err := io.Copy(outFile, file); err != nil {
		return "", fmt.Errorf("failed to save file: %v", err)
	}

	return filePath, nil
}

// CleanupExpiredFiles periodically checks and deletes expired files
func CleanupExpiredFiles(conn *pgx.Conn) {
    for {
        query := `SELECT id, file_name FROM files WHERE upload_date < NOW() - INTERVAL '30 days'`
        rows, err := conn.Query(context.Background(), query)
        if err != nil {
            fmt.Printf("Error querying expired files: %v\n", err)
            continue
        }

        for rows.Next() {
            var fileID int
            var fileName string
            if err := rows.Scan(&fileID, &fileName); err != nil {
                fmt.Printf("Error scanning row: %v\n", err)
                continue
            }

            filePath := filepath.Join(UploadDir, fileName)
            if err := os.Remove(filePath); err != nil {
                fmt.Printf("Error deleting file %s: %v\n", filePath, err)
                continue
            }

            _, err := conn.Exec(context.Background(), `DELETE FROM files WHERE id = $1`, fileID)
            if err != nil {
                fmt.Printf("Error deleting file metadata from database: %v\n", err)
            }
        }

        if err := rows.Err(); err != nil {
            fmt.Printf("Error in rows iteration: %v\n", err)
        }

        time.Sleep(CheckInterval)
    }
}


