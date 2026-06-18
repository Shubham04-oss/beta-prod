package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/pubsub"
)

func main() {
	ctx := context.Background()
	projectID := "demo-synq"

	os.Setenv("PUBSUB_EMULATOR_HOST", "shubhams-mac-mini.local:8085")

	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	topicID := "pim-events"
	topic := client.Topic(topicID)

	exists, err := topic.Exists(ctx)
	if err != nil {
		log.Fatalf("Failed to check if topic exists: %v", err)
	}

	if !exists {
		_, err = client.CreateTopic(ctx, topicID)
		if err != nil {
			log.Fatalf("Failed to create topic: %v", err)
		}
		fmt.Printf("Topic %s created successfully.\n", topicID)
	} else {
		fmt.Printf("Topic %s already exists.\n", topicID)
	}

	// Create subscriptions
	subscriptions := []string{"unified-pim-events", "procurement-pim-events"}
	for _, subID := range subscriptions {
		sub := client.Subscription(subID)
		exists, err := sub.Exists(ctx)
		if err != nil {
			log.Fatalf("Failed to check subscription %s: %v", subID, err)
		}
		if !exists {
			_, err = client.CreateSubscription(ctx, subID, pubsub.SubscriptionConfig{
				Topic:       topic,
				AckDeadline: 20 * 1000 * 1000 * 1000, // 20 seconds
			})
			if err != nil {
				log.Fatalf("Failed to create subscription %s: %v", subID, err)
			}
			fmt.Printf("Subscription %s created successfully.\n", subID)
		} else {
			fmt.Printf("Subscription %s already exists.\n", subID)
		}
	}
}
