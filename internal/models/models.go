package models

import "time"

// User представляет пользователя в системе
type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

// Habit представляет привычку
type Habit struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Color       string    `json:"color"`
	CreatedAt   time.Time `json:"created_at"`

	IsCompletedToday bool `json:"is_completed_today"`
	CurrentStreak    int  `json:"current_streak"`
}

// HabitLog представляет запись о выполнении
type HabitLog struct {
	ID            int       `json:"id"`
	HabitID       int       `json:"habit_id"`
	CompletedDate time.Time `json:"completed_date"`
	CreatedAt     time.Time `json:"created_at"`
}

// Stats структура для страницы статистики
type HabitStats struct {
	HabitID       int    `json:"habit_id"`
	Title         string `json:"title"`
	TotalDays     int    `json:"total_days"`
	CurrentStreak int    `json:"current_streak"`
	LongestStreak int    `json:"longest_streak"`
}
