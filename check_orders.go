package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbpool, err := pgxpool.New(context.Background(), "postgres://postgres:postgres@localhost:5432/synq?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	rows, err := dbpool.Query(context.Background(), "SELECT column_name FROM information_schema.columns WHERE table_name='orders'")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var col string
		rows.Scan(&col)
		fmt.Println(col)
	}
}
