package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq" // PostgreSQL driver
)

func connectDB() (*sql.DB, error) {
	db, err := sql.Open("postgres", "user=db_user password=example dbname=guardian_db sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}

func Init() {
	var err error
	db, err = connectDB()
	if err != nil {
		log.Fatalf("Database initialization error: %v", err)
	}
}
