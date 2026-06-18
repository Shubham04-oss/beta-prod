package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"encoding/json"

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
	
    res, err := sdk.Unified.ListUnifiedConnections(ctx, operations.ListUnifiedConnectionsRequest{})
    if err != nil {
        log.Fatalf("Error getting connections: %v", err)
    }

    if len(res.Connections) == 0 {
        log.Println("No connections found.")
        return
    }

	for _, conn := range res.Connections {
		fmt.Printf("Connection: %v Env: %v\n", *conn.ID, *conn.Environment)

		// Fetch Accounting Orders
		ordersRes, err := sdk.Accounting.ListAccountingOrders(ctx, operations.ListAccountingOrdersRequest{
			ConnectionID: *conn.ID,
		})
		if err != nil {
			log.Printf("Error fetching orders for conn %v: %v", *conn.ID, err)
			continue
		}

		fmt.Printf("Found %d orders for connection %v\n", len(ordersRes.AccountingOrders), *conn.ID)
		if len(ordersRes.AccountingOrders) > 0 {
			b, _ := json.MarshalIndent(ordersRes.AccountingOrders[0], "", "  ")
			fmt.Printf("Sample Order:\n%s\n", string(b))
		}
	}
}
