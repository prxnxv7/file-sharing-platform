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
    httpSwagger "github.com/swaggo/http-swagger/v2"
    // "github.com/swaggo/http-swagger"
    "github.com/gorilla/mux"
	_ "github.com/lib/pq"
	_ "file-sharing-platform/docs"

)

// @title File Sharing Platform API
// @version 1.0
// @description This is a file sharing platform API documentation.
// @host localhost:8080
// @BasePath /
func main() {
    db, err := config.ConnectDB()
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close(context.Background())

	services.InitS3()

    go services.CleanupExpiredFiles(db)
	go services.InitRedis()

	hub := services.NewHub()
	go hub.RunHub()
    
    r := mux.NewRouter()

    r.HandleFunc("/register", handlers.RegisterUser).Methods("POST")
    r.HandleFunc("/login", handlers.LoginUser).Methods("POST")

	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		services.ServeWs(hub, w, r)
	})
    r.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	)).Methods(http.MethodGet)

    protected := r.PathPrefix("/").Subrouter()
    protected.Use(middleware.AuthMiddleware)
    protected.Use(middleware.RateLimiterMiddleware)
    protected.HandleFunc("/file/{id:[0-9]+}", handlers.GetFile).Methods("GET")
    protected.HandleFunc("/search", handlers.SearchFiles).Methods("GET")
    protected.HandleFunc("/upload/{user_id:[0-9]+}", handlers.UploadFile).Methods("POST")

    log.Fatal(http.ListenAndServe(":8080", r))
}
