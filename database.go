package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq" // PostgreSQL driver
)

func connectDB() (*sql.DB, error) {
	db, err := sql.Open("postgres", environments.DatabaseConn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}

func initDatabase() (*sql.DB, error) {
	var err error
	db, err = connectDB()
	if err != nil {
		log.Fatalf("Database initialization error: %v", err)
		return nil, err
	}
	return db, nil
}
