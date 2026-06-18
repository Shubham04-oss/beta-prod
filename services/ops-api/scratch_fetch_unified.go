package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	unified "github.com/unified-to/unified-go-sdk"
	"github.com/unified-to/unified-go-sdk/pkg/models/operations"
)

func main() {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("No .env found, using env vars")
	}
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env found, using env vars")
	}

	token := os.Getenv("UNIFIED_TO_TOKEN")
	if token == "" {
		log.Fatal("UNIFIED_TO_TOKEN is not set")
	}

	sdk := unified.New(
		unified.WithSecurity(token),
	)

	ctx := context.Background()

	// Try to list orders from Commerce API across all connections?
	// Unified requires ConnectionID to list orders.
	// Let's get connections from Unified API
	res, err := sdk.Unified.ListUnifiedConnections(ctx, operations.ListUnifiedConnectionsRequest{})
	if err != nil {
		log.Fatalf("Error getting connections: %v", err)
	}

	if len(res.Connections) == 0 {
		log.Println("No connections found.")
		return
	}

	for _, conn := range res.Connections {
		fmt.Printf("Connection: %v (%v)\n", *conn.ID, conn.Environment)

		if conn.Environment != nil && *conn.Environment != "Sandbox" {
			continue
		}

		// Fetch orders
		ordersRes, err := sdk.Commerce.ListCommerceOrders(ctx, operations.ListCommerceOrdersRequest{
			ConnectionID: *conn.ID,
		})
		if err != nil {
			log.Printf("Error fetching orders for conn %v: %v", *conn.ID, err)
			continue
		}

		fmt.Printf("Found %d orders for connection %v\n", len(ordersRes.CommerceOrders), *conn.ID)
		if len(ordersRes.CommerceOrders) > 0 {
			b, _ := json.MarshalIndent(ordersRes.CommerceOrders[0], "", "  ")
			fmt.Printf("Sample Order:\n%s\n", string(b))
		}
	}
}
