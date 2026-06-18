package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	dbpool, err := pgxpool.New(ctx, "postgres://dev:dev@shubhams-mac-mini.local:5432/synq_db?sslmode=disable")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	sqlBytes, err := os.ReadFile("infrastructure/postgres/inventory.sql")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read sql file: %v\n", err)
		os.Exit(1)
	}

	_, err = dbpool.Exec(ctx, string(sqlBytes))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to execute sql: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Migration successful")
}
