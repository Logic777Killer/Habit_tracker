package handlers

import (
	"database/sql"
	"encoding/json"
	"habit-tracker/internal/auth"
	"habit-tracker/internal/database"
	"habit-tracker/internal/models"
	"log"
	"net/http"
	"strings"
)

// RegisterRequest структура входящего JSON
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest структура входящего JSON
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterHandler обрабатывает регистрацию
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Хэшируем пароль
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Вставляем в БД
	query := "INSERT INTO users (username, email, password_hash, role) VALUES ($1, $2, $3, 'user') RETURNING id"
	var userID int
	err = database.DB.QueryRow(query, req.Username, req.Email, hashedPassword).Scan(&userID)

	if err != nil {
		log.Printf("DB Error during registration: %v", err)

		if strings.Contains(err.Error(), "duplicate key") {
			http.Error(w, "User with this email or username already exists", http.StatusConflict)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User created successfully"})
}

// LoginHandler обрабатывает вход
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Ищем пользователя по email
	var user models.User
	query := "SELECT id, username, email, password_hash, role FROM users WHERE email = $1"
	err := database.DB.QueryRow(query, req.Email).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Проверяем пароль
	if !auth.CheckPasswordHash(req.Password, user.PasswordHash) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Генерируем токен
	token, err := auth.GenerateToken(user.ID, user.Role)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Отправляем токен
	json.NewEncoder(w).Encode(map[string]string{"token": token, "role": user.Role})
}
