package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	unified "github.com/unified-to/unified-go-sdk"
	"github.com/unified-to/unified-go-sdk/pkg/models/shared"
)

func main() {
    _ = godotenv.Load("../../.env")
	token := os.Getenv("UNIFIED_TO_TOKEN")
	sdk := unified.New(unified.WithSecurity(token))
	ctx := context.Background()

    // Create a Sandbox connection for Shopify (commerce)
    provider := "shopify"
    env := "Sandbox"
    conn := shared.Connection{
        Categories: []shared.PropertyConnectionCategories{shared.PropertyConnectionCategoriesCommerce},
        Provider:   &provider,
        Environment: &env,
    }

	res, err := sdk.Unified.CreateUnifiedConnection(ctx, &conn)
	if err != nil {
		log.Fatalf("Error creating connection: %v", err)
	}

	fmt.Printf("Successfully created Sandbox Connection ID: %v\n", *res.Connection.ID)
}
