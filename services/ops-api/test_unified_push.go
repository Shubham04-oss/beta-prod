package main

import (
	"context"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/synq/ops-api/internal/unified"
	"github.com/synq/pkg/db"
)

func main() {
	_ = godotenv.Load(".env")
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL must be set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// 1. Get an existing tenant and product
	var tenantID pgtype.UUID
	err = pool.QueryRow(ctx, "SELECT id FROM tenants LIMIT 1").Scan(&tenantID)
	if err != nil {
		log.Fatalf("No tenants found: %v", err)
	}

	var productID, orgID pgtype.UUID
	err = pool.QueryRow(ctx, "SELECT id, org_id FROM products WHERE tenant_id = $1 LIMIT 1", tenantID).Scan(&productID, &orgID)
	if err != nil {
		log.Fatalf("No products found: %v", err)
	}

	// 2. Insert dummy connection
	queries := db.New(pool)
	queries.CreateCommerceConnection(ctx, db.CreateCommerceConnectionParams{
		OrgID:               orgID,
		TenantID:            tenantID,
		UnifiedConnectionID: "fake_shopify_123",
		Provider:            "shopify",
	})

	log.Printf("Inserted mock connection for tenant %s", uuid.UUID(tenantID.Bytes).String())

	// 3. Init Service and Trigger Push
	svc := unified.NewService(pool, "dummy_token")

	err = svc.ProcessPush(ctx, uuid.UUID(tenantID.Bytes).String(), uuid.UUID(orgID.Bytes).String(), uuid.UUID(productID.Bytes).String(), "UPSERT")
	if err != nil {
		log.Printf("ProcessPush returned error (expected since token is fake): %v", err)
	} else {
		log.Printf("ProcessPush succeeded!")
	}
}
