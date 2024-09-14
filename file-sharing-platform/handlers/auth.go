package handlers

import (
	"context"
	"encoding/json"
	"file-sharing-platform/config"
	"file-sharing-platform/models"
	"file-sharing-platform/utils"
	"net/http"

	// "github.com/jackc/pgx/v4"
)

func RegisterUser(w http.ResponseWriter, r *http.Request) {
    var user models.User
    // Decode the JSON body
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Hash password
    hashedPassword, _ := utils.HashPassword(user.Password)
    user.Password = hashedPassword

    conn, _ := config.ConnectDB()
    defer conn.Close(context.Background())

    // Insert user into database
    _, err := conn.Exec(context.Background(), "INSERT INTO users (email, password) VALUES ($1, $2)", user.Email, user.Password)
    if err != nil {
        http.Error(w, "Unable to create user", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"message": "User created successfully"})
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
    var user models.User
    json.NewDecoder(r.Body).Decode(&user)

    conn, _ := config.ConnectDB()
    defer conn.Close(context.Background())

    // Check if user exists
    var hashedPassword string
    err := conn.QueryRow(context.Background(), "SELECT password FROM users WHERE email=$1", user.Email).Scan(&hashedPassword)
    if err != nil {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

    // Compare hashed password
    if !utils.CheckPasswordHash(user.Password, hashedPassword) {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

    // Generate JWT token
    token, _ := utils.GenerateJWT(user.Email)
    json.NewEncoder(w).Encode(map[string]string{"token": token})
}
