package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbUrl := "postgres://dev:dev@shubhams-mac-mini.local:5432/synq_db"
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer pool.Close()

	sqlBytes, err := os.ReadFile("infrastructure/postgres/pim_taxonomy.sql")
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	_, err = pool.Exec(ctx, string(sqlBytes))
	if err != nil {
		log.Fatalf("Failed to execute migration: %v", err)
	}

	fmt.Println("Migration successful!")
}
