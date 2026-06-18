package main

import (
	"context"
	"fmt"
	"log"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/google/uuid"
)

func main() {
	dbpool, err := pgxpool.New(context.Background(), "postgres://dev:dev@shubhams-mac-mini.local:5432/synq_db?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	tenantID := "bd103a09-90d6-4a3a-8749-cc41bdd592a7"
	orgID := "9acbf3ff-663c-49ab-ae50-fa6d97568cba"
	productID := uuid.New().String()

	_, err = dbpool.Exec(context.Background(), `
		INSERT INTO products (id, org_id, tenant_id, title)
		VALUES ($1, $2, $3, $4)
	`, productID, orgID, tenantID, "Test Product")

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted product successfully!")
}
