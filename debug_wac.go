package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	dbpool, err := pgxpool.New(ctx, "postgres://dev:dev@shubhams-mac-mini.local:5432/synq_db?sslmode=disable")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	var id pgtype.UUID
	var transactionType string
	var unitCost pgtype.Numeric
	err = dbpool.QueryRow(ctx, "SELECT id, transaction_type, unit_cost FROM inventory_ledger ORDER BY created_at DESC LIMIT 1").Scan(&id, &transactionType, &unitCost)
	if err != nil {
		fmt.Println(err)
		return
	}
    
    val, _ := unitCost.Float64Value()
	fmt.Printf("Latest Ledger -> ID: %s | Type: %s | UnitCost: %v\n", id.Bytes, transactionType, val.Float64)
}
