package main

import (
	"fmt"
	"habit-tracker/internal/auth"
	"habit-tracker/internal/config"
	"habit-tracker/internal/database"
	"habit-tracker/internal/handlers"
	"habit-tracker/internal/middleware"
	"log"
	"net/http"
)

func main() {
	cfg := config.LoadConfig()

	if err := auth.SetJWTSecret(cfg.JWTSecret); err != nil {
		log.Fatalf("JWT configuration error: %v", err)
	}

	if err := database.InitDB(cfg); err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	defer database.DB.Close()

	setupRoutes()

	fmt.Printf("Server starting on port %s...\n", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}

func setupRoutes() {
	fs := http.FileServer(http.Dir("./web"))
	http.Handle("/", fs)

	http.HandleFunc("/api/register", handlers.RegisterHandler)
	http.HandleFunc("/api/login", handlers.LoginHandler)

	userRoute := func(handler http.HandlerFunc) http.HandlerFunc {
		return middleware.AuthMiddleware(middleware.RequireRole("user", "admin")(handler))
	}

	http.HandleFunc("/api/habits", userRoute(handlers.GetHabitsHandler))
	http.HandleFunc("/api/habits/create", userRoute(handlers.CreateHabitHandler))

	http.HandleFunc("/api/habits/toggle", userRoute(handlers.ToggleHabitHandler))

	http.HandleFunc("/api/habits/logs", userRoute(handlers.GetHabitLogsHandler))

	http.HandleFunc("/api/habits/stats", userRoute(handlers.GetHabitStatsHandler))

	http.HandleFunc("/api/habits/delete", userRoute(handlers.DeleteHabitHandler))

}
