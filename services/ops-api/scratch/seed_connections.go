package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load("/Users/shubham/Projects/synq/services/ops-api/.env")
	dbURL := os.Getenv("DATABASE_URL")
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	if len(os.Args) > 1 && os.Args[1] == "revert" {
		_, err := pool.Exec(ctx, "DELETE FROM commerce_item_mappings WHERE connection_id IN (SELECT id FROM commerce_connections WHERE unified_connection_id LIKE 'mock_stress_connection_%')")
		if err != nil {
			log.Fatalf("Revert Mappings failed: %v\n", err)
		}
		
		_, err = pool.Exec(ctx, "DELETE FROM sync_failures_dlq WHERE connection_id IN (SELECT id FROM commerce_connections WHERE unified_connection_id LIKE 'mock_stress_connection_%')")
		if err != nil {
			log.Fatalf("Revert DLQ failed: %v\n", err)
		}

		tag, err := pool.Exec(ctx, "DELETE FROM commerce_connections WHERE unified_connection_id LIKE 'mock_stress_connection_%'")
		if err != nil {
			log.Fatalf("Revert Connections failed: %v\n", err)
		}
		fmt.Printf("Reverted %d dummy connections and all their mappings!\n", tag.RowsAffected())
		return
	}

	query := `
		INSERT INTO commerce_connections (tenant_id, org_id, unified_connection_id, status)
		SELECT t.id, o.id, 'mock_stress_connection_' || substr(md5(random()::text), 1, 6), 'ACTIVE'
		FROM tenants t
		JOIN organizations o ON o.tenant_id = t.id
		WHERE NOT EXISTS (
			SELECT 1 FROM commerce_connections c WHERE c.tenant_id = t.id AND c.status = 'ACTIVE'
		);
	`
	tag, err := pool.Exec(ctx, query)
	if err != nil {
		log.Fatalf("Insert failed: %v\n", err)
	}
	fmt.Printf("Inserted %d dummy connections for stress test!\n", tag.RowsAffected())
}
