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
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	// 1. Get a random location
	var locID, orgID, tenantID string
	err = dbpool.QueryRow(ctx, "SELECT id, org_id, tenant_id FROM locations LIMIT 1").Scan(&locID, &orgID, &tenantID)
	if err != nil {
		fmt.Printf("No locations found: %v\n", err)
		// Insert a default location
		err = dbpool.QueryRow(ctx, "SELECT org_id, tenant_id FROM products LIMIT 1").Scan(&orgID, &tenantID)
		if err != nil {
			fmt.Printf("No products found, skipping seeding\n")
			return
		}
		err = dbpool.QueryRow(ctx, "INSERT INTO locations (org_id, tenant_id, name) VALUES ($1, $2, 'Main Warehouse') RETURNING id", orgID, tenantID).Scan(&locID)
		if err != nil {
			fmt.Printf("Failed to insert location: %v\n", err)
			return
		}
	}

	// 2. Fetch up to 10 variants
	rows, err := dbpool.Query(ctx, "SELECT id FROM product_variants WHERE tenant_id = $1 LIMIT 10", tenantID)
	if err != nil {
		fmt.Printf("Error fetching variants: %v\n", err)
		return
	}
	defer rows.Close()

	var variantIDs []string
	for rows.Next() {
		var vid string
		rows.Scan(&vid)
		variantIDs = append(variantIDs, vid)
	}

	if len(variantIDs) == 0 {
		fmt.Println("No variants to seed.")
		return
	}

	// 3. Update cost_price and insert ledger
	quantities := []int{15, 120, 0, 5, 45, 8, 300, 2}
	for i, vid := range variantIDs {
		price := float64(20 + i*10)
		_, err = dbpool.Exec(ctx, "UPDATE product_variants SET cost_price = $1 WHERE id = $2", price, vid)
		if err != nil {
			fmt.Printf("Failed to update price: %v\n", err)
		}

		q := quantities[i%len(quantities)]
		_, err = dbpool.Exec(ctx, `
			INSERT INTO inventory_ledger 
			(org_id, tenant_id, variant_id, location_id, transaction_type, quantity_delta, notes) 
			VALUES ($1, $2, $3, $4, 'RESTOCK', $5, 'Initial seeding')`, 
			orgID, tenantID, vid, locID, q)
		if err != nil {
			fmt.Printf("Failed to insert ledger for variant %s: %v\n", vid, err)
		}
	}

	fmt.Println("Seeding complete.")
}
