package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v4"
	"github.com/joho/godotenv"
)

var (
    S3Client *s3.Client
    CheckInterval = 24 * time.Hour
)

func InitS3() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	s3Cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(os.Getenv("AWS_REGION")),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"),
			"",
		)),
	)
	if err != nil {
		fmt.Printf("Unable to load AWS config: %v", err)
		return
	}

	S3Client = s3.NewFromConfig(s3Cfg)
	fmt.Println("S3 client initialized successfully!")
}

func UploadFile(file multipart.File, fileHeader *multipart.FileHeader, userID int) (string, error) {
	buffer := new(bytes.Buffer)
	if _, err := io.Copy(buffer, file); err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	s3Path := fmt.Sprintf("uploads/user-%d/%s", userID, fileHeader.Filename)

	_, err := S3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:             aws.String(os.Getenv("S3_BUCKET_NAME")),
		Key:                aws.String(s3Path),
		ACL:                "public-read",
		Body:               bytes.NewReader(buffer.Bytes()),
		ContentLength:      aws.Int64(int64(buffer.Len())),
		ContentType:        aws.String(http.DetectContentType(buffer.Bytes())),
		ContentDisposition: aws.String("attachment"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %v", err)
	}

	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", os.Getenv("S3_BUCKET_NAME"), os.Getenv("AWS_REGION"), s3Path), nil
}

func CleanupExpiredFiles(conn *pgx.Conn) {
	for {
		query := `SELECT id, s3_url FROM files WHERE upload_date < NOW() - INTERVAL '30 days'`
		rows, err := conn.Query(context.Background(), query)
		if err != nil {
			fmt.Printf("Error querying expired files: %v\n", err)
			continue
		}

		for rows.Next() {
			var fileID int
			var s3URL string
			if err := rows.Scan(&fileID, &s3URL); err != nil {
				fmt.Printf("Error scanning row: %v\n", err)
				continue
			}

			s3Path := s3URL[strings.LastIndex(s3URL, "/")+1:]

			_, err = S3Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
				Bucket: aws.String(os.Getenv("S3_BUCKET_NAME")),
				Key:    aws.String(s3Path),
			})
			if err != nil {
				fmt.Printf("Error deleting S3 object: %v\n", err)
			}

			_, err = conn.Exec(context.Background(), `DELETE FROM files WHERE id = $1`, fileID)
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


