package clients

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

// NewValkeyClient initializes and returns a connection client to Valkey (Redis-compatible)
func NewValkeyClient() (*redis.Client, error) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = os.Getenv("VALKEY_URL")
	}
	if redisURL == "" {
		// Default fallback for local developer running compose stack
		redisURL = "redis://localhost:6379"
	}

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	// Adjust options for beta/production workloads
	opts.DialTimeout = 5 * time.Second
	opts.ReadTimeout = 3 * time.Second
	opts.WriteTimeout = 3 * time.Second
	opts.PoolSize = 10

	client := redis.NewClient(opts)

	// Verify the connection works
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Valkey/Redis connection check failed: %v", err)
	} else {
		log.Printf("Successfully established connection to Valkey/Redis at %s", opts.Addr)
	}

	return client, nil
}
