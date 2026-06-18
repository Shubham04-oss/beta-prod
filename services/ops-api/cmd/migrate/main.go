package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbURL := "postgres://dev:dev@shubhams-mac-mini.local:5432/synq_db"
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	sqlBytes, err := os.ReadFile("../../infrastructure/postgres/unified.sql")
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
	}
	sql := string(sqlBytes)

	_, err = pool.Exec(context.Background(), sql)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	fmt.Println("Migration successful!")
}
