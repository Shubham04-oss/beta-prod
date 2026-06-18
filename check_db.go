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

	var count int
	err = dbpool.QueryRow(context.Background(), "SELECT count(*) FROM organizations").Scan(&count)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Organizations count: %d\n", count)
	
	rows, _ := dbpool.Query(context.Background(), "SELECT id FROM organizations LIMIT 1")
	defer rows.Close()
	for rows.Next() {
		var id string
		rows.Scan(&id)
		fmt.Printf("Org ID: %s\n", id)
	}
	
	var countTenants int
	dbpool.QueryRow(context.Background(), "SELECT count(*) FROM tenants").Scan(&countTenants)
	fmt.Printf("Tenants count: %d\n", countTenants)

	rowsT, _ := dbpool.Query(context.Background(), "SELECT id FROM tenants LIMIT 1")
	defer rowsT.Close()
	for rowsT.Next() {
		var id string
		rowsT.Scan(&id)
		fmt.Printf("Tenant ID: %s\n", id)
	}
}
