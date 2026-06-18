package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dsn := "postgres://dev:dev@shubhams-mac-mini.local:5432/synq_db"
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	sqlBytes, err := os.ReadFile("../../infrastructure/postgres/pim_advanced.sql")
	if err != nil {
		log.Fatalf("Unable to read sql file: %v\n", err)
	}

	_, err = pool.Exec(context.Background(), string(sqlBytes))
	if err != nil {
		log.Fatalf("Execution failed: %v", err)
	}

	fmt.Println("Successfully created pim_advanced schema!")
}

