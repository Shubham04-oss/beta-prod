package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	sdk "github.com/unified-to/unified-go-sdk"
	"github.com/unified-to/unified-go-sdk/pkg/models/shared"
)

func main() {
	godotenv.Load(".env")
	token := os.Getenv("UNIFIED_TO_TOKEN")
	workspace := os.Getenv("UNIFIED_WORKSPACE_ID")

	client := sdk.New(
		sdk.WithSecurity(token),
	)

	ctx := context.Background()

	// Create a sandbox connection
	conn, err := client.Unified.CreateUnifiedConnection(ctx, shared.Connection{
		IntegrationType: "shopify",
		Environment:     sdk.String("Sandbox"),
		WorkspaceID:     &workspace,
		Categories: []shared.PropertyConnectionCategories{
			shared.PropertyConnectionCategoriesCommerce,
		},
		Permissions: []shared.PropertyConnectionPermissions{
			shared.PropertyConnectionPermissionsAccountingOrderRead,
			shared.PropertyConnectionPermissionsAccountingOrderWrite,
		},
	})

	if err != nil {
		log.Fatalf("Error creating sandbox connection: %v", err)
	}

	fmt.Printf("Created sandbox connection: %s\n", *conn.Connection.ID)
}
