package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/google/uuid"
)

func main() {
	dbpool, err := pgxpool.New(context.Background(), "postgres://dev:dev@192.168.1.6:5432/synq_db?sslmode=disable")
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}
	defer dbpool.Close()

	tenantIDStr := "bd103a09-90d6-4a3a-8749-cc41bdd592a7"
	orgIDStr := "9acbf3ff-663c-49ab-ae50-fa6d97568cba"
	unifiedConnID := "6a31c60957acafe1c4104eac"
	
	id := uuid.New()

	_, err = dbpool.Exec(context.Background(), `
		INSERT INTO commerce_connections (id, tenant_id, org_id, unified_connection_id, provider, status)
		VALUES ($1, $2, $3, $4, 'shopify', 'ACTIVE')
	`, id, tenantIDStr, orgIDStr, unifiedConnID)

	if err != nil {
		log.Fatalf("Insert failed: %v", err)
	}
	fmt.Printf("Successfully registered connection %s for tenant %s\n", unifiedConnID, tenantIDStr)
}
