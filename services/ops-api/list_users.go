package main

import (
	"context"
	"fmt"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/iterator"
)

func main() {
	os.Setenv("FIREBASE_AUTH_EMULATOR_HOST", "shubhams-mac-mini.local:9099")
	os.Setenv("GCLOUD_PROJECT", "demo-synq")

	ctx := context.Background()
	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		log.Fatalf("Error initializing firebase app: %v", err)
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		log.Fatalf("Error getting Auth client: %v", err)
	}

	iter := authClient.Users(ctx, "")
	count := 0
	for {
		u, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Error listing users: %v", err)
		}
		fmt.Printf("User: %s (Tenant: %s)\n", u.Email, u.TenantID)
		count++
	}
	fmt.Printf("Total users: %d\n", count)
}
