package handlers

import (
	"encoding/json"
	"github.com/lib/pq"
	"habit-tracker/internal/database"
	"habit-tracker/internal/middleware"
	"habit-tracker/internal/models"
	"log"
	"net/http"
	"strconv"
	"time"
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

	var req struct {
		HabitID int    `json:"habit_id"`
		Date    string `json:"date"` // Опционально: YYYY-MM-DD
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Определяем целевую дату
	targetDate := time.Now().Format("2006-01-02")
	if req.Date != "" {
		if _, err := time.Parse("2006-01-02", req.Date); err != nil {
			http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		targetDate = req.Date
	}

	// Проверка принадлежности привычки
	var ownerID int
	err := database.DB.QueryRow("SELECT user_id FROM habits WHERE id = $1", req.HabitID).Scan(&ownerID)
	if err != nil || ownerID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Проверяем, есть ли запись на эту дату
	var exists bool
	err = database.DB.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM habit_logs WHERE habit_id = $1 AND completed_date = $2::date)`,
		req.HabitID, targetDate,
	).Scan(&exists)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Переключаем статус
	if exists {
		_, err = database.DB.Exec(
			`DELETE FROM habit_logs WHERE habit_id = $1 AND completed_date = $2::date`,
			req.HabitID, targetDate,
		)
	} else {
		_, err = database.DB.Exec(
			`INSERT INTO habit_logs (habit_id, completed_date) VALUES ($1, $2::date)`,
			req.HabitID, targetDate,
		)
	}

	if err != nil {
		http.Error(w, "Failed to update", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"completed": !exists, "date": targetDate})
}

func GetHabitLogsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	habitID, _ := strconv.Atoi(r.URL.Query().Get("habit_id"))
	month, _ := strconv.Atoi(r.URL.Query().Get("month"))
	year, _ := strconv.Atoi(r.URL.Query().Get("year"))

	if habitID == 0 || month == 0 || year == 0 {
		http.Error(w, "Missing habit_id, month or year", http.StatusBadRequest)
		return
	}

	rows, err := database.DB.Query(`
		SELECT completed_date FROM habit_logs 
		WHERE habit_id = $1 AND EXTRACT(YEAR FROM completed_date) = $2 AND EXTRACT(MONTH FROM completed_date) = $3
	`, habitID, year, month)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var dates []string
	for rows.Next() {
		var date time.Time
		if err := rows.Scan(&date); err == nil {
			dates = append(dates, date.Format("2006-01-02"))
		}
	}

	json.NewEncoder(w).Encode(map[string][]string{"dates": dates})
}

func GetHabitStatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rows, err := database.DB.Query(`
		SELECT h.id, h.title, COUNT(hl.completed_date) as total_days,
		       ARRAY(SELECT completed_date::text FROM habit_logs WHERE habit_id = h.id ORDER BY completed_date DESC) as dates
		FROM habits h
		LEFT JOIN habit_logs hl ON h.id = hl.habit_id
		WHERE h.user_id = $1
		GROUP BY h.id, h.title
	`, userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var stats []map[string]interface{}
	for rows.Next() {
		var id, totalDays int
		var title string
		var dateStrs []string

		if err := rows.Scan(&id, &title, &totalDays, pq.Array(&dateStrs)); err != nil {
			log.Printf("Stats scan error: %v", err)
			continue
		}

		var dates []time.Time
		for _, ds := range dateStrs {
			if t, err := time.Parse("2006-01-02", ds); err == nil {
				dates = append(dates, t)
			}
		}

		stats = append(stats, map[string]interface{}{
			"id":             id,
			"title":          title,
			"total_days":     totalDays,
			"current_streak": calculateCurrentStreak(dates),
			"longest_streak": calculateLongestStreak(dates),
		})
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Rows iteration error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(stats)
}

// calculateCurrentStreak считает дни подряд от сегодня
func calculateCurrentStreak(dates []time.Time) int {
	if len(dates) == 0 {
		return 0
	}

	today := time.Now().Truncate(24 * time.Hour)
	streak := 0
	checkDate := today

	for _, d := range dates {
		date := d.Truncate(24 * time.Hour)
		if date.Equal(checkDate) {
			streak++
			checkDate = checkDate.AddDate(0, 0, -1)
		} else if date.Before(checkDate) {
			break
		}
	}
	return streak
}

// calculateLongestStreak считает самый длинный стрик в истории
func calculateLongestStreak(dates []time.Time) int {
	if len(dates) == 0 {
		return 0
	}

	maxStreak := 1
	currentStreak := 1

	for i := 1; i < len(dates); i++ {
		prev := dates[i-1].Truncate(24 * time.Hour)
		curr := dates[i].Truncate(24 * time.Hour)

		if prev.Sub(curr) == 24*time.Hour {
			currentStreak++
			if currentStreak > maxStreak {
				maxStreak = currentStreak
			}
		} else {
			currentStreak = 1
		}
	}
	return maxStreak
}

func DeleteHabitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		HabitID int `json:"habit_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.HabitID == 0 {
		http.Error(w, "Invalid habit_id", http.StatusBadRequest)
		return
	}

	var ownerID int
	err := database.DB.QueryRow("SELECT user_id FROM habits WHERE id = $1", req.HabitID).Scan(&ownerID)
	if err != nil || ownerID != userID {
		http.Error(w, "Forbidden or not found", http.StatusForbidden)
		return
	}

	_, err = database.DB.Exec("DELETE FROM habits WHERE id = $1", req.HabitID)
	if err != nil {
		http.Error(w, "Failed to delete", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Habit deleted"})
}
