package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	dbpool, err := pgxpool.New(ctx, "postgres://dev:dev@shubhams-mac-mini.local:5432/synq_db?sslmode=disable")
	if err != nil {
		fmt.Printf("Unable to connect: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()

    var orgID, tenantID string
	dbpool.QueryRow(ctx, "SELECT org_id, tenant_id FROM tenants LIMIT 1").Scan(&orgID, &tenantID)
    
    var count int
    dbpool.QueryRow(ctx, "SELECT count(*) FROM locations WHERE tenant_id = $1", tenantID).Scan(&count)
    if count == 0 {
        _, err = dbpool.Exec(ctx, "INSERT INTO locations (org_id, tenant_id, name, type) VALUES ($1, $2, 'Main Warehouse', 'WAREHOUSE')", orgID, tenantID)
        fmt.Println("Created location: ", err)
    }
}
