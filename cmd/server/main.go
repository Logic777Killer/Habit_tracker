package main

import (
	"fmt"
	"habit-tracker/internal/config"
	"habit-tracker/internal/database"
	"habit-tracker/internal/handlers"
	"habit-tracker/internal/middleware"
	"log"
	"net/http"
)

func main() {
	cfg := config.LoadConfig()

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

	http.HandleFunc("/api/habits", middleware.AuthMiddleware(handlers.GetHabitsHandler))
	http.HandleFunc("/api/habits/create", middleware.AuthMiddleware(handlers.CreateHabitHandler))

	http.HandleFunc("/api/habits/toggle", middleware.AuthMiddleware(handlers.ToggleHabitHandler))

}
