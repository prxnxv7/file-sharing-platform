package main

import (
    "log"
    "net/http"
    "file-sharing-platform/handlers"
    "file-sharing-platform/middleware"
    "github.com/gorilla/mux"
)

func main() {
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
    protected.HandleFunc("/files/{id:[0-9]+}", handlers.GetFile).Methods("GET")
    protected.HandleFunc("/upload/{user_id:[0-9]+}", handlers.UploadFile).Methods("POST")

    log.Fatal(http.ListenAndServe(":8080", r))
}
