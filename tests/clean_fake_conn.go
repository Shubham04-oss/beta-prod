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

	_, err = dbpool.Exec(context.Background(), "DELETE FROM commerce_connections WHERE unified_connection_id = 'fake_shopify_123'")
	if err != nil {
		log.Fatalf("Delete failed: %v", err)
	}
	fmt.Println("Cleaned up fake_shopify_123 connection")
}
