package main

import (
	"context"
	"io/ioutil"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbURL := "postgres://dev:dev@shubhams-mac-mini.local:5432/synq_db"
	dbpool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbpool.Close()

	sqlBytes, err := ioutil.ReadFile("../../infrastructure/postgres/channels.sql")
	if err != nil {
		log.Fatalf("Unable to read sql file: %v\n", err)
	}

	_, err = dbpool.Exec(context.Background(), string(sqlBytes))
	if err != nil {
		log.Fatalf("Failed to execute migration: %v\n", err)
	}

	log.Println("Migration successful!")
}
