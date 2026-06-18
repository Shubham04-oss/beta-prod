package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"cloud.google.com/go/storage"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/synq/ops-api/internal/ucp"
	"google.golang.org/api/option"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found or error loading it")
	}

	ctx := context.Background()

	// 1. Connect to Postgres
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://dev:dev@localhost:5432/synq_db?sslmode=disable"
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// 2. Connect to GCS Emulator
	// We point the storage client to the emulator running on port 4443
	err = os.Setenv("STORAGE_EMULATOR_HOST", "shubhams-mac-mini.local:4443")
	if err != nil {
		log.Fatalf("Failed to set env: %v", err)
	}

	client, err := storage.NewClient(ctx, option.WithoutAuthentication(), option.WithEndpoint("http://shubhams-mac-mini.local:4443/storage/v1/"))
	if err != nil {
		log.Fatalf("Failed to create storage client: %v", err)
	}
	defer client.Close()

	// 3. Initialize Activities
	activities := ucp.NewActivities(pool, client)

	// 4. Run Extract Catalog
	// Query for a tenant that actually has products
	var tenantID string
	err = pool.QueryRow(ctx, "SELECT tenant_id FROM products LIMIT 1").Scan(&tenantID)
	if err != nil {
		log.Fatalf("Failed to get tenant ID with products (run test_db.go first): %v", err)
	}
	log.Printf("Using Tenant ID: %s", tenantID)

	feed, err := activities.ExtractCatalogActivity(ctx, tenantID)
	if err != nil {
		log.Fatalf("ExtractCatalogActivity failed: %v", err)
	}

	b, _ := json.MarshalIndent(feed, "", "  ")
	log.Printf("Extracted Feed: \n%s\n", string(b))

	// 5. Upload to GCS
	url, err := activities.UploadFeedToGCSActivity(ctx, feed, tenantID)
	if err != nil {
		log.Fatalf("UploadFeedToGCSActivity failed: %v", err)
	}

	log.Printf("UCP Feed successfully generated and uploaded to: %s", url)
}
