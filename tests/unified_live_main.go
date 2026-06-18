package main

import (
	"context"
	"fmt"
	"log"

	unified "github.com/unified-to/unified-go-sdk"
	"github.com/unified-to/unified-go-sdk/pkg/models/operations"
)

func main() {
	sdk := unified.New(unified.WithSecurity("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiI2YTMwNDMyZDI1MDc0YmExZmE5NDBjYTkiLCJ3b3Jrc3BhY2VfaWQiOiI2YTMwNDMyZDI1MDc0YmExZmE5NDBjYWUiLCJpYXQiOjE3ODE1NDc4MjEsImtleV9pZCI6IjZhMzA0MzJlMjUwNzRiYTFmYTk0MGNiMyIsIm5vbmNlIjoidG40ajZiNjZoTW9OY3JlUlFkY3FQTktyMUNOaXZyN3MifQ.ixFL9DbwHTQazOnccjhn6ut1pD3voR3X8JQ95EDua3I"))

	ctx := context.Background()
	res, err := sdk.Unified.ListUnifiedConnections(ctx, operations.ListUnifiedConnectionsRequest{})
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	
	if res.Connections != nil {
		fmt.Printf("Found %d connections\n", len(res.Connections))
		for _, conn := range res.Connections {
			fmt.Printf("Connection: ID=%s, Integration=%s\n", *conn.ID, conn.IntegrationType)
		}
	} else {
		fmt.Println("No connections found.")
	}
}
