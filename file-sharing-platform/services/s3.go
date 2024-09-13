package services

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"database/sql"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	_ "github.com/lib/pq"
)

var (
	S3Region = os.Getenv("AWS_REGION")
	S3Bucket = os.Getenv("AWS_BUCKET")
	DBConn   *sql.DB
)

var svc *s3.S3

func init() {
	var err error
	DBConn, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
}

func UploadToS3(file multipart.File, fileHeader *multipart.FileHeader, userID int) (string, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(S3Region),
	})
	if err != nil {
		return "", err
	}

	svc := s3.New(sess)

	buffer := make([]byte, fileHeader.Size)
	file.Read(buffer)

	s3Path := fmt.Sprintf("uploads/user-%d/%s", userID, fileHeader.Filename)

	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket:             aws.String(S3Bucket),
		Key:                aws.String(s3Path),
		ACL:                aws.String("public-read"),
		Body:               bytes.NewReader(buffer),
		ContentLength:      aws.Int64(fileHeader.Size),
		ContentType:        aws.String(http.DetectContentType(buffer)),
		ContentDisposition: aws.String("attachment"),
	})

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", S3Bucket, S3Region, s3Path), nil
}

func CleanUpExpiredFiles() {
	ctx := context.Background()

	// Example: Get files older than 30 days
	threshold := time.Now().Add(-30 * 24 * time.Hour)
	rows, err := DBConn.QueryContext(ctx, "SELECT id, s3_url FROM files WHERE upload_date < $1", threshold)
	if err != nil {
		log.Printf("Failed to query files for cleanup: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var s3URL string
		if err := rows.Scan(&id, &s3URL); err != nil {
			log.Printf("Failed to scan file row: %v", err)
			continue
		}

		// Delete from S3
		s3Path := s3URL[strings.LastIndex(s3URL, "/")+1:]
		_, err = svc.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(S3Bucket),
			Key:    aws.String(s3Path),
		})
		if err != nil {
			log.Printf("Failed to delete S3 object: %v", err)
		}

		// Delete from PostgreSQL
		_, err = DBConn.ExecContext(ctx, "DELETE FROM files WHERE id = $1", id)
		if err != nil {
			log.Printf("Failed to delete file from database: %v", err)
		}
	}
}
