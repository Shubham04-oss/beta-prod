package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/synq/ops-api/internal/unified"
)

func main() {
	dbpool, err := pgxpool.New(context.Background(), "postgres://dev:dev@192.168.1.6:5432/synq_db?sslmode=disable")
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}
	defer dbpool.Close()

	svc := unified.NewService(dbpool, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiI2YTMwNDMyZDI1MDc0YmExZmE5NDBjYTkiLCJ3b3Jrc3BhY2VfaWQiOiI2YTMwNDMyZDI1MDc0YmExZmE5NDBjYWUiLCJpYXQiOjE3ODE1NDc4MjEsImtleV9pZCI6IjZhMzA0MzJlMjUwNzRiYTFmYTk0MGNiMyIsIm5vbmNlIjoidG40ajZiNjZoTW9OY3JlUlFkY3FQTktyMUNOaXZyN3MifQ.ixFL9DbwHTQazOnccjhn6ut1pD3voR3X8JQ95EDua3I")
	
	err = svc.ProcessPush(
		context.Background(),
		"bd103a09-90d6-4a3a-8749-cc41bdd592a7", // tenantID
		"9acbf3ff-663c-49ab-ae50-fa6d97568cba", // orgID
		"661bc97c-5b3d-4b64-8597-afc9426dc582", // productID
		"UPSERT",
	)
	
	if err != nil {
		log.Fatalf("Push failed: %v", err)
	}
	fmt.Println("Successfully pushed product to Unified.to Sandbox!")
}
