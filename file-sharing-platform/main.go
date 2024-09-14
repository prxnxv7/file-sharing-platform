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

    // Initialize the Redis client for rate-limiting and caching
	go services.InitRedis()

	// Initialize WebSocket hub and run it in a goroutine
	hub := services.NewHub()
	go hub.RunHub()
    
    r := mux.NewRouter()

    // Routes
    r.HandleFunc("/register", handlers.RegisterUser).Methods("POST")
    r.HandleFunc("/login", handlers.LoginUser).Methods("POST")

	// WebSocket route for file upload notifications
	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		services.ServeWs(hub, w, r)
	})

    // Apply authentication middleware to protected routes
    protected := r.PathPrefix("/").Subrouter()
    protected.Use(middleware.AuthMiddleware)
    protected.Use(middleware.RateLimiterMiddleware)
    protected.HandleFunc("/file/{id:[0-9]+}", handlers.GetFile).Methods("GET")
    protected.HandleFunc("/search}", handlers.SearchFiles).Methods("GET")
    protected.HandleFunc("/upload/{user_id:[0-9]+}", handlers.UploadFile).Methods("POST")

    log.Fatal(http.ListenAndServe(":8080", r))
}
