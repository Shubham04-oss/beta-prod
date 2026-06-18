package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load("/Users/shubham/Projects/synq/services/ops-api/.env")
	dbURL := os.Getenv("DATABASE_URL")
	ctx := context.Background()

	pool, _ := pgxpool.New(ctx, dbURL)
	defer pool.Close()

	tx, _ := pool.Begin(ctx)
	defer tx.Rollback(ctx)

	var orgID string
	err := tx.QueryRow(ctx, "INSERT INTO organizations (name) VALUES ('Test Org') RETURNING id").Scan(&orgID)
	if err != nil {
		fmt.Printf("Org Error: %v\n", err)
		return
	}

	var tenantID string
	err = tx.QueryRow(ctx, "INSERT INTO tenants (org_id, name) VALUES ($1, 'Test Tenant') RETURNING id", orgID).Scan(&tenantID)
	if err != nil {
		fmt.Printf("Tenant Error: %v\n", err)
		return
	}

	_, err = tx.Exec(ctx, "INSERT INTO commerce_connections (org_id, tenant_id, unified_connection_id, provider, status) VALUES ($1, $2, $3, 'shopify', 'ACTIVE')", orgID, tenantID, "mock_stress_connection_"+tenantID)
	if err != nil {
		fmt.Printf("Commerce Connection Error: %v\n", err)
		return
	}

	fmt.Println("Success!")
}
