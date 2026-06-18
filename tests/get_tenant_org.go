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

	var tenantID string
	err = dbpool.QueryRow(context.Background(), "SELECT id FROM tenants LIMIT 1").Scan(&tenantID)
	if err != nil {
		log.Fatalf("Select tenant failed: %v", err)
	}
	fmt.Printf("Tenant ID: %s\n", tenantID)

	var orgID string
	err = dbpool.QueryRow(context.Background(), "SELECT id FROM organizations LIMIT 1").Scan(&orgID)
	if err != nil {
		log.Fatalf("Select org failed: %v", err)
	}
	fmt.Printf("Org ID: %s\n", orgID)
}
