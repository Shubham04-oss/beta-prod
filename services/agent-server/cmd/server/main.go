package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/synq/agent-server/internal/core"
	"github.com/synq/agent-server/internal/llm"
	"github.com/synq/agent-server/internal/telemetry"
	"github.com/synq/agent-server/internal/tools"
)

func main() {
	log.Println("Starting Synq Agent Server...")

	ctx := context.Background()

	// Initialize Postgres Connection Pool
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://dev:dev@localhost:5432/synq_db"
	}
	dbpool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbpool.Close()

	// 0. Initialize ADK Telemetry
	shutdownTelemetry, err := telemetry.InitADKTelemetry(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize ADK Telemetry: %v", err)
	}
	defer func() {
		if err := shutdownTelemetry(ctx); err != nil {
			log.Printf("Telemetry shutdown error: %v", err)
		}
	}()

	// 1. Initialize Vertex AI Model Router
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	location := os.Getenv("GOOGLE_CLOUD_LOCATION")
	if projectID == "" {
		projectID = "dummy-dev-project"
		location = "us-central1"
	}
	
	_, err = llm.SetupVertexRouter(ctx, projectID, location)
	if err != nil {
		log.Fatalf("Failed to initialize Vertex AI Router: %v", err)
	}

	// 2. Initialize Temporal Client
	temporalMgr, err := core.NewTemporalManager("localhost:7233")
	if err != nil {
		log.Fatalf("Failed to initialize Temporal Client: %v", err)
	}
	defer temporalMgr.Close()

	// 2. Initialize Worker for our Agent Task Queue
	agentWorker := temporalMgr.NewWorker("AGENT_TASK_QUEUE")

	// 3. Register Unified.to Activities
	unifiedActivities := tools.NewUnifiedActivities(dbpool, "dummy_unified_workspace_id")
	agentWorker.RegisterActivity(unifiedActivities)

	// 4. Start the Worker in a goroutine
	go func() {
		log.Println("Worker listening on AGENT_TASK_QUEUE...")
		if err := agentWorker.Start(); err != nil {
			log.Fatalf("Failed to start Temporal worker: %v", err)
		}
	}()
	defer agentWorker.Stop()

	// 5. Wait for graceful shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down Agent Server gracefully...")
}
