package handlers

import (
	"bytes"
	"encoding/json"
	"habit-tracker/internal/database"
	"habit-tracker/internal/middleware"
	"habit-tracker/internal/models"
	"io"
	"log"
	"net/http"
)

// CreateHabitHandler создает новую привычку
func CreateHabitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var habit struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Color       string `json:"color"`
	}
	if err := json.NewDecoder(r.Body).Decode(&habit); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	query := "INSERT INTO habits (user_id, title, description, color) VALUES ($1, $2, $3, $4) RETURNING id"
	var id int
	err := database.DB.QueryRow(query, userID, habit.Title, habit.Description, habit.Color).Scan(&id)
	if err != nil {
		http.Error(w, "Failed to create habit", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "message": "Habit created"})
}

// GetHabitsHandler возвращает список привычек пользователя
func GetHabitsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Запрос с подзапросом для проверки: выполнена ли привычка сегодня
	query := `
		SELECT h.id, h.title, h.description, h.color, h.created_at,
		       EXISTS(
		           SELECT 1 FROM habit_logs 
		           WHERE habit_id = h.id 
		           AND completed_date = CURRENT_DATE
		       ) as is_completed_today
		FROM habits h
		WHERE h.user_id = $1
		ORDER BY h.created_at DESC
	`
	rows, err := database.DB.Query(query, userID)
	if err != nil {
		http.Error(w, "Failed to fetch habits", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var habits []models.Habit
	for rows.Next() {
		var h models.Habit
		err := rows.Scan(&h.ID, &h.Title, &h.Description, &h.Color, &h.CreatedAt, &h.IsCompletedToday)
		if err != nil {
			continue
		}
		habits = append(habits, h)
	}

	json.NewEncoder(w).Encode(habits)
}

// ToggleHabitHandler переключает статус выполнения привычки на сегодня
func ToggleHabitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Читаем тело запроса для отладки
	bodyBytes, _ := io.ReadAll(r.Body)
	log.Printf("Toggle request body: %s", string(bodyBytes)) // <-- Лог для отладки

	// Восстанавливаем тело, чтобы его можно было прочитать ещё раз
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var req struct {
		HabitID int `json:"habit_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("JSON decode error: %v", err) // <-- Лог ошибки парсинга
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.HabitID == 0 {
		log.Printf("HabitID is zero or missing")
		http.Error(w, "habit_id is required", http.StatusBadRequest)
		return
	}

	// Проверяем, что привычка принадлежит текущему пользователю
	var ownerID int
	err := database.DB.QueryRow("SELECT user_id FROM habits WHERE id = $1", req.HabitID).Scan(&ownerID)
	if err != nil || ownerID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// ... остальная логика переключения (без изменений) ...
	var exists bool
	err = database.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM habit_logs WHERE habit_id = $1 AND completed_date = CURRENT_DATE)`, req.HabitID).Scan(&exists)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	if exists {
		_, err = database.DB.Exec("DELETE FROM habit_logs WHERE habit_id = $1 AND completed_date = CURRENT_DATE", req.HabitID)
	} else {
		_, err = database.DB.Exec("INSERT INTO habit_logs (habit_id, completed_date) VALUES ($1, CURRENT_DATE)", req.HabitID)
	}

	if err != nil {
		http.Error(w, "Failed to update", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]bool{"completed": !exists})
}
