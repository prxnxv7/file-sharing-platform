# File Sharing Platform

A file sharing platform built using Go, PostgreSQL, Redis, and AWS S3. This application allows users to upload, download, and manage files, with features such as rate limiting, caching, and periodic file cleanup.

## Tech Stack

- **Go**: Programming language used for the backend logic.
- **PostgreSQL**: Database for storing file metadata and user information.
- **Redis**: Used for caching file metadata and rate limiting.
- **AWS S3**: Used for storing uploaded files.
- **Docker**: Containerization of the application.
- **Swagger**: API documentation and testing interface.
- **Gorilla Mux**: Router for handling HTTP requests.
- **AWS SDK for Go v2**: AWS SDK for interacting with S3.

## Features

- **File Upload**: Users can upload files to the server, which are then stored in AWS S3.
- **File Download**: Users can download files that they have uploaded.
- **File Metadata Caching**: File metadata is cached in Redis for fast access.
- **Rate Limiting**: API rate limiting is enforced using Redis.
- **Periodic Cleanup**: Expired files are periodically cleaned up from both the database and AWS S3.
- **Swagger Documentation**: API documentation and testing are available via Swagger.
- **JWT Authentication**: Authentication is featured by JSON Web Tokens.

### Prerequisites

- Docker
- Docker Compose
- AWS Account (for S3 and environment variables)

### Environment Variables

Created an `.env` file in the root directory with the following variables:


## Setup and Usage

1. **Clone the repository:**
    ```bash
    git clone <repository-url>
    cd <repository-directory>
    ```

2. **Build and run the Docker containers:**
    ```bash
    docker-compose up --build
    ```

3. **Configure Environment Variables:**
   - Create a `.env` file in the root directory with the following variables:
     ```env
     AWS_REGION=<your-aws-region>
     AWS_ACCESS_KEY_ID=<your-access-key-id>
     AWS_SECRET_ACCESS_KEY=<your-secret-access-key>
     S3_BUCKET_NAME=<your-s3-bucket-name>
     REDIS_ADDR=<your-redis-address>
     DB_USER=<your-db-user>
     DB_PASSWORD=<your-db-password>
     DB_NAME=<your-db-name>
     ```

4. **Access the API:**
   - The API runs on `http://localhost:8080`.

5. **API Documentation:**
   - Access Swagger documentation at `http://localhost:8080/swagger/index.html`.


## API Endpoints

- **Register User**
  - `POST /register` - Register a new user.

- **Login User**
  - `POST /login` - Log in a user.

- **Get File**
  - `GET /file/{id:[0-9]+}` - Retrieve file details by file ID.

- **Search Files**
  - `GET /search` - Search for files based on query parameters.

- **Upload File**
  - `POST /upload/{user_id:[0-9]+}` - Upload a file.

- **WebSocket for File Upload Notifications**
  - `GET /ws` - WebSocket connection for receiving file upload notifications.

