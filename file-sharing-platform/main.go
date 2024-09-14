package main

import (
	// "database/sql"
	"context"
	"file-sharing-platform/config"
	"file-sharing-platform/handlers"
	"file-sharing-platform/middleware"
	"file-sharing-platform/services"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
    // Connect to the database
    db, err := config.ConnectDB()
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close(context.Background())

    // Start the file cleanup service
    go services.CleanupExpiredFiles(db)
    
    r := mux.NewRouter()

    // Routes
    r.HandleFunc("/register", handlers.RegisterUser).Methods("POST")
    r.HandleFunc("/login", handlers.LoginUser).Methods("POST")
    r.HandleFunc("/upload/{user_id:[0-9]+}", handlers.UploadFile).Methods("POST")
    r.HandleFunc("/file/{id:[0-9]+}", handlers.GetFile).Methods("GET")
    r.HandleFunc("/search", handlers.SearchFiles).Methods("GET")

    // Apply authentication middleware to protected routes
    protected := r.PathPrefix("/").Subrouter()
    protected.Use(middleware.AuthMiddleware)
    protected.HandleFunc("/file/{id:[0-9]+}", handlers.GetFile).Methods("GET")
    protected.HandleFunc("/search}", handlers.SearchFiles).Methods("GET")
    protected.HandleFunc("/upload/{user_id:[0-9]+}", handlers.UploadFile).Methods("POST")

    log.Fatal(http.ListenAndServe(":8080", r))
}
