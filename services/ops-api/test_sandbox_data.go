package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/unified-to/unified-go-sdk/pkg/models/operations"
	sdk "github.com/unified-to/unified-go-sdk"
)

func main() {
	godotenv.Load(".env")
	token := os.Getenv("UNIFIED_TO_TOKEN")
	if token == "" {
		log.Fatal("UNIFIED_TO_TOKEN is required")
	}

	client := sdk.New(
		sdk.WithSecurity(token),
	)

	ctx := context.Background()

	res, err := client.Unified.ListUnifiedConnections(ctx, operations.ListUnifiedConnectionsRequest{
		Env: sdk.String("Sandbox"),
	})
	if err != nil {
		log.Fatalf("Error listing connections: %v", err)
	}

	fmt.Printf("Found %d connections\n", len(res.Connections))
	for _, conn := range res.Connections {
		fmt.Printf("Connection: %s (%s) - Env: %v\n", *conn.ID, conn.IntegrationType, *conn.Environment)
		
		// List accounting orders
		ordersRes, err := client.Accounting.ListAccountingOrders(ctx, operations.ListAccountingOrdersRequest{
			ConnectionID: *conn.ID,
		})
		if err != nil {
			fmt.Printf("  Error fetching orders: %v\n", err)
			continue
		}

		fmt.Printf("  Found %d orders\n", len(ordersRes.AccountingOrders))
		for _, order := range ordersRes.AccountingOrders {
			fmt.Printf("    Order ID: %s, Status: %v\n", *order.ID, order.Status)
		}
	}
}
