package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"habit-tracker/internal/config"
)

var DB *sql.DB

func InitDB(cfg *config.Config) error {
	var err error

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("error opening db: %w", err)
	}
	
	err = DB.Ping()
	if err != nil {
		return fmt.Errorf("error connecting to db: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL")
	return nil
}
