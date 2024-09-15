package handlers

import (
	"context"
	"encoding/json"
	"file-sharing-platform/config"
	"file-sharing-platform/models"
	"file-sharing-platform/utils"
	"log"
	"net/http"
)
// @Summary Register a new user
// @Description Registers a new user with email and password
// @Tags user
// @Accept json
// @Produce json
// @Param user body models.User true "User Registration"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /register [post]
func RegisterUser(w http.ResponseWriter, r *http.Request) {
    log.Printf("Received request: %s %s", r.Method, r.URL.Path)
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Printf("RegisterUser: Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		log.Printf("RegisterUser: Error hashing password: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	user.Password = hashedPassword

	conn, err := config.ConnectDB()
	if err != nil {
		log.Printf("RegisterUser: Error connecting to database: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer conn.Close(context.Background())

	_, err = conn.Exec(context.Background(), "INSERT INTO users (email, password) VALUES ($1, $2)", user.Email, user.Password)
	if err != nil {
		log.Printf("RegisterUser: Error inserting user into database: %v", err)
		http.Error(w, "Unable to create user", http.StatusInternalServerError)
		return
	}

	log.Printf("RegisterUser: User created successfully with email %s", user.Email)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User created successfully"})
}

// @Summary Login a user
// @Description Authenticates a user and returns a JWT token
// @Tags user
// @Accept json
// @Produce json
// @Param user body models.User true "User Login"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /login [post]
func LoginUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Printf("LoginUser: Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	conn, err := config.ConnectDB()
	if err != nil {
		log.Printf("LoginUser: Error connecting to database: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer conn.Close(context.Background())

	var hashedPassword string
	err = conn.QueryRow(context.Background(), "SELECT password FROM users WHERE email=$1", user.Email).Scan(&hashedPassword)
	if err != nil {
		log.Printf("LoginUser: Error querying database for email %s: %v", user.Email, err)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if !utils.CheckPasswordHash(user.Password, hashedPassword) {
		log.Printf("LoginUser: Invalid credentials for email %s", user.Email)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := utils.GenerateJWT(user.Email)
	if err != nil {
		log.Printf("LoginUser: Error generating JWT: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("LoginUser: User logged in successfully with email %s", user.Email)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
