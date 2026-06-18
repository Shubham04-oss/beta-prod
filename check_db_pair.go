package main

import (
	"context"
	"fmt"
	"log"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbpool, err := pgxpool.New(context.Background(), "postgres://dev:dev@shubhams-mac-mini.local:5432/synq_db?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	var orgID string
	var tenantID string
	err = dbpool.QueryRow(context.Background(), "SELECT id, tenant_id FROM organizations LIMIT 1").Scan(&orgID, &tenantID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("MATCHING PAIR:\nOrgID: %s\nTenantID: %s\n", orgID, tenantID)
}
