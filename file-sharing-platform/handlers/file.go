package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"file-sharing-platform/config"
	"file-sharing-platform/services"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gorilla/mux"
)

// UploadFile uploads a file to S3 and stores metadata in the database and cache
// @Summary Upload a file
// @Description Uploads a file to an S3 bucket and stores its metadata in a database and cache
// @Tags files
// @Accept multipart/form-data
// @Produce json
// @Param user_id path int true "User ID"
// @Param file formData file true "File to upload"
// @Success 200 {string} string "File upload started"
// @Failure 400 {string} string "Invalid user ID or error reading file"
// @Failure 500 {string} string "Internal server error"
// @Router /upload/{user_id} [post]
func UploadFile(w http.ResponseWriter, r *http.Request) {
    log.Println("Starting UploadFile handler") 

    userID, err := strconv.Atoi(mux.Vars(r)["user_id"])
    if err != nil {
        log.Printf("Invalid user ID: %v\n", err)
        http.Error(w, "Invalid user ID", http.StatusBadRequest)
        return
    }

    log.Printf("Processing upload for user ID: %d\n", userID)

    file, fileHeader, err := r.FormFile("file")
    if err != nil {
        log.Printf("Error reading file: %v\n", err)
        http.Error(w, "Error reading file", http.StatusBadRequest)
        return
    }
    defer file.Close()

    log.Printf("Received file: %s\n", fileHeader.Filename) 

    resultChan := make(chan error)

	go func() {
        log.Println("Starting file processing goroutine") 
		s3Client := services.S3Client

		bucketName := os.Getenv("S3_BUCKET_NAME")
		s3Key := "uploads/" + fileHeader.Filename

        log.Printf("Uploading file to S3: %s\n", s3Key)
        _, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
            Bucket: aws.String(bucketName),
            Key:    aws.String(s3Key),
            Body:   file,
        })
		if err != nil {
			log.Printf("Error uploading file to S3: %v\n", err)
			resultChan <- err
			return
		}

        log.Println("File uploaded to S3 successfully")

		db, err := config.ConnectDB()
		if err != nil {
			log.Printf("Error connecting to DB: %v\n", err) 
			resultChan <- err
			return
		}
		defer db.Close(context.Background())

		query := `INSERT INTO files (user_id, file_name, file_size, s3_key) VALUES ($1, $2, $3, $4)`
		_, err = db.Exec(context.Background(), query, userID, fileHeader.Filename, fileHeader.Size, s3Key)
		if err != nil {
			log.Printf("Error inserting file metadata into DB: %v\n", err) 
			resultChan <- err
			return
		}

        log.Println("File metadata saved to DB") 

		fileMetadata := map[string]interface{}{
			"user_id":   userID,
			"file_name": fileHeader.Filename,
			"file_size": fileHeader.Size,
			"s3_key":    s3Key,
		}
		metadataJSON, _ := json.Marshal(fileMetadata)
		cacheKey := "file_metadata:" + fileHeader.Filename
		cacheErr := services.CacheFileMetadata(cacheKey, string(metadataJSON), 5*time.Minute)
		if cacheErr != nil {
			log.Printf("Error caching file metadata: %v\n", cacheErr)
			resultChan <- cacheErr
			return
		}

        log.Println("File metadata cached successfully") 

		resultChan <- nil
		uploadCompleteChan <- true
	}()

    log.Println("Responding to client: File upload started") 
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("File upload started"))

    if err := <-resultChan; err != nil {
        log.Printf("Error occurred in file upload process: %v\n", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        uploadCompleteChan <- false
    }
}

// GetFile retrieves a file from S3 based on file ID
// @Summary Retrieve a file
// @Description Retrieves a file from S3 based on its ID and serves it to the client
// @Tags files
// @Produce application/octet-stream
// @Param file_id path int true "File ID"
// @Success 200 {file} string "File content"
// @Failure 400 {string} string "Invalid file ID"
// @Failure 404 {string} string "File not found"
// @Failure 500 {string} string "Internal server error"
// @Router /files/{file_id} [get]
func GetFile(w http.ResponseWriter, r *http.Request) {
	fileID, err := strconv.Atoi(mux.Vars(r)["file_id"])
	if err != nil {
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	cacheKey := "file_metadata:" + strconv.Itoa(fileID)
	cachedMetadata, cacheErr := services.GetCachedFileMetadata(cacheKey)
	var s3Key string
	if cacheErr == nil && cachedMetadata != "" {
		var metadata map[string]interface{}
		json.Unmarshal([]byte(cachedMetadata), &metadata)
		s3Key = metadata["s3_key"].(string)
	} else {
		db, err := config.ConnectDB()
		if err != nil {
			http.Error(w, "Error connecting to database", http.StatusInternalServerError)
			return
		}
		defer db.Close(context.Background())

		query := `SELECT file_name, file_size, s3_key FROM files WHERE id = $1`
		row := db.QueryRow(context.Background(), query, fileID)

		var fileName string
		var fileSize int64
		err = row.Scan(&fileName, &fileSize, &s3Key)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "File not found", http.StatusNotFound)
			} else {
				http.Error(w, "Error retrieving file", http.StatusInternalServerError)
			}
			return
		}

		fileMetadata := map[string]interface{}{
			"file_name": fileName,
			"file_size": fileSize,
			"s3_key":    s3Key,
		}
		metadataJSON, _ := json.Marshal(fileMetadata)
		cacheErr = services.CacheFileMetadata(cacheKey, string(metadataJSON), 5*time.Minute)
		if cacheErr != nil {
			http.Error(w, "Error caching file metadata", http.StatusInternalServerError)
			return
		}
	}

	s3Client := services.S3Client

	output, err := s3Client.GetObject(context.TODO(),&s3.GetObjectInput{
		Bucket: aws.String(os.Getenv("S3_BUCKET_NAME")),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		http.Error(w, "File does not exist", http.StatusNotFound)
		return
	}
	defer output.Body.Close()

	w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(s3Key))
	w.Header().Set("Content-Type", *output.ContentType)
	io.Copy(w, output.Body)
}
