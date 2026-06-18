package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbpool, err := pgxpool.New(context.Background(), "postgres://dev:dev@192.168.1.6:5432/synq_db?sslmode=disable")
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}
	defer dbpool.Close()

	var productID string
	err = dbpool.QueryRow(context.Background(), "SELECT id FROM products WHERE tenant_id = 'bd103a09-90d6-4a3a-8749-cc41bdd592a7' LIMIT 1").Scan(&productID)
	if err != nil {
		log.Fatalf("Select product failed: %v", err)
	}
	fmt.Printf("Product ID: %s\n", productID)
}
