package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	gpubsub "cloud.google.com/go/pubsub"
	"crypto/tls"
	"firebase.google.com/go/v4"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/synq/ops-api/internal/api"
	"github.com/synq/ops-api/internal/clients"
	"github.com/synq/ops-api/internal/importexport"
	auth_middleware "github.com/synq/ops-api/internal/middleware"
	"github.com/synq/ops-api/internal/oms"
	"github.com/synq/ops-api/internal/pim"
	"github.com/synq/ops-api/internal/procurement"
	"github.com/synq/ops-api/internal/pubsub"
	"github.com/synq/ops-api/internal/service"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/synq/ops-api/internal/telemetry"
	"github.com/synq/ops-api/internal/unified"
	"github.com/synq/pkg/authcontext"
	"github.com/synq/pkg/db"
	"github.com/synq/pkg/events"

	"github.com/exaring/otelpgx"
	"github.com/riandyrn/otelchi"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system environment variables")
	}

	ctx := context.Background()
	port := "8080"

	// 0. Initialize Telemetry
	otlpEndpoint := os.Getenv("OTLP_ENDPOINT")
	shutdown, err := telemetry.SetupTracing(ctx, "ops-api", otlpEndpoint)
	if err != nil {
		log.Fatalf("Failed to setup telemetry: %v", err)
	}
	defer shutdown(ctx)

	// 1. Initialize Postgres Connection Pool
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}
	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		log.Fatalf("Failed to parse db config: %v", err)
	}
	config.ConnConfig.Tracer = otelpgx.NewTracer()
	dbpool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbpool.Close()

	// 1.5. Initialize Valkey (Redis-compatible) Connection
	valkeyClient, err := clients.NewValkeyClient()
	if err != nil {
		log.Fatalf("Failed to initialize Valkey client: %v", err)
	}
	defer valkeyClient.Close()

	// 2. Initialize Firebase Admin SDK
	firebaseHost := os.Getenv("FIREBASE_AUTH_EMULATOR_HOST")
	if firebaseHost != "" {
		if err := os.Setenv("FIREBASE_AUTH_EMULATOR_HOST", firebaseHost); err != nil {
			log.Printf("Warning: failed to set FIREBASE_AUTH_EMULATOR_HOST: %v", err)
		}
	}

	gcloudProject := firstNonEmpty(os.Getenv("GCLOUD_PROJECT"), os.Getenv("GOOGLE_CLOUD_PROJECT"))
	if gcloudProject == "" {
		log.Fatal("GCLOUD_PROJECT or GOOGLE_CLOUD_PROJECT is required")
	}
	if err := os.Setenv("GCLOUD_PROJECT", gcloudProject); err != nil {
		log.Printf("Warning: failed to set GCLOUD_PROJECT: %v", err)
	}

	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		log.Fatalf("Error initializing firebase app: %v", err)
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		log.Fatalf("Error getting Auth client: %v", err)
	}

	// 3. Initialize Services and Handlers
	queries := db.New(dbpool)
	lifecycleService := service.NewLifecycleService(dbpool, authClient)
	tenantHandler := api.NewTenantHandler(lifecycleService)
	unifiedWorkspaceID := requiredEnv("UNIFIED_WORKSPACE_ID")
	integrationsHandler := api.NewIntegrationsHandler(dbpool, unifiedWorkspaceID)
	organizationHandler := api.NewOrganizationHandlers(queries, lifecycleService)
	auditHandler := api.NewAuditHandlers(queries)
	settingsHandler := api.NewSettingsHandlers(queries)
	channelHandler := api.NewChannelHandlers(queries)
	teamWorkspaceHandler := api.NewTeamWorkspaceHandlers(dbpool)
	ticketStore := service.NewWSTicketStore(queries)
	websocketHandler := api.NewWebSocketHandler(ticketStore)
	// Create PIM Service & Handler
	publisher, err := events.NewPubSubPublisher(ctx, gcloudProject)
	if err != nil {
		log.Printf("Failed to init pubsub publisher: %v", err)
	}
	pimService := pim.NewService(dbpool, publisher)
	pimHandler := api.NewPIMHandlers(pimService, queries)
	inventoryHandler := api.NewInventoryHandlers(queries)

	// Create Data Manipulator and Eventarc Handler
	importExportManipulator := importexport.NewManipulator(dbpool)
	eventarcHandler, _ := importexport.NewEventarcHandler(importExportManipulator)

	unifiedToken := requiredEnv("UNIFIED_TO_TOKEN")
	unifiedService := unified.NewService(dbpool, unifiedToken)
	unifiedService.StartWorkerPool(ctx)
	defer unifiedService.StopWorkerPool()

	unifiedWebhookSecret := os.Getenv("UNIFIED_WEBHOOK_SECRET")
	unifiedHandler := api.NewUnifiedHandler(dbpool, unifiedService, unifiedWebhookSecret)

	// Procurement Service
	importProcurement := true // Placeholder for the actual import
	_ = importProcurement

	subscriber, err := events.NewPubSubSubscriber(ctx, gcloudProject)
	if err == nil {
		go func() {
			err := unified.StartUnifiedEventSubscriber(context.Background(), subscriber, unifiedService, "unified-pim-events")
			if err != nil {
				log.Printf("Unified subscriber failed: %v", err)
			}
		}()

		// Wire up Procurement Auto-PO Worker
		procurementSvc := procurement.NewService(dbpool)
		go func() {
			err := procurement.StartProcurementEventSubscriber(context.Background(), subscriber, procurementSvc, "procurement-pim-events")
			if err != nil {
				log.Printf("Procurement subscriber failed: %v", err)
			}
		}()
	} else {
		log.Printf("Failed to init pubsub subscriber: %v", err)
	}

	// Initialize GCP Pub/Sub Client for Outbox Forwarder
	gcpPubsubClient, err := gpubsub.NewClient(ctx, gcloudProject)
	if err != nil {
		log.Printf("Failed to create GCP PubSub client: %v", err)
	} else {
		outboxForwarder := pubsub.NewOutboxForwarder(dbpool, gcpPubsubClient)
		outboxForwarder.Start(ctx)
		log.Println("Started Real-Time Outbox Forwarder (LISTEN/NOTIFY)")
	}

	// 4. Configure Router
	r := chi.NewRouter()

	// Add OpenTelemetry instrumentation
	r.Use(otelchi.Middleware("ops-api", otelchi.WithChiRoutes(r)))

	// Basic CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "https://app.synq.com"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Limit request body size to 10MB
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
			next.ServeHTTP(w, r)
		})
	})

	// Initialize Valkey/Redis-backed sliding window rate limiter (e.g., max 100 requests per minute)
	limiter := auth_middleware.NewRateLimiter(valkeyClient, 100, 1*time.Minute)
	r.Use(limiter.Handler)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Initialize Temporal Client early so public webhooks can use it
	temporalHost := os.Getenv("TEMPORAL_HOST_PORT")
	if temporalHost == "" {
		temporalHost = "localhost:7233"
	}
	temporalClient, err := client.Dial(client.Options{
		HostPort: temporalHost,
		// TODO: Configure proper certs in production
		ConnectionOptions: client.ConnectionOptions{
			TLS: &tls.Config{InsecureSkipVerify: os.Getenv("ENV") != "production"},
		},
	})
	if err != nil {
		log.Printf("Failed to create Temporal client: %v", err)
	} else {
		defer temporalClient.Close()
		// Initialize Unified Webhook Handler with Temporal
		webhookHandler := unified.NewWebhookHandler(temporalClient, dbpool, unifiedWebhookSecret)
		webhookHandler.RegisterRoutes(r)
	}

	// (Auth Middleware moved to protected group)

	// Public Routes
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Unauthenticated onboarding route
	tenantHandler.RegisterRoutes(r)

	// Mock Inventory Adjust for k6 test
	r.Post("/api/v1/inventory/adjust", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Eventarc Webhook (Public, secured by Eventarc OIDC/IAM natively)
	r.Post("/api/v1/pim/import/webhook", eventarcHandler.HandleImportWebhook)

	// UCP Discovery route (it enforces its own auth checking the API key)
	ucpHandler := api.NewUCPHandler()
	ucpHandler.RegisterRoutes(r)

	// Register Prometheus Metrics endpoint
	telemetry.SetupMetrics(r)

	if err == nil {
		// Initialize Temporal Worker
		w := worker.New(temporalClient, "oms-task-queue-v3", worker.Options{})

		// Register OMS Activities and Workflows
		omsRepo := oms.NewRepository(dbpool)
		omsActivities := oms.NewActivities(omsRepo, unifiedService)
		w.RegisterActivity(omsActivities.CreateOrderActivity)
		w.RegisterActivity(omsActivities.ReserveInventoryActivity)
		w.RegisterActivity(omsActivities.AuthorizePaymentActivity)
		w.RegisterActivity(omsActivities.ConfirmOrderActivity)
		w.RegisterActivity(omsActivities.EmitOrderPlacedActivity)
		w.RegisterActivity(omsActivities.ReleaseInventoryActivity)
		w.RegisterActivity(omsActivities.MarkOrderFulfilledActivity)
		w.RegisterActivity(omsActivities.MarkReturnRequestedActivity)
		w.RegisterActivity(omsActivities.SyncFulfillmentToChannelActivity)

		unifiedActivities := unified.NewActivities(temporalClient, unifiedService, dbpool)
		w.RegisterActivity(unifiedActivities.PollCommerceOrdersActivity)
		w.RegisterActivity(unifiedActivities.MapInboundOrderActivity)

		w.RegisterWorkflow(oms.OrderCreationWorkflow)
		w.RegisterWorkflow(oms.OrderFulfillmentWorkflow)
		w.RegisterWorkflow(oms.OrderReturnWorkflow)
		w.RegisterWorkflow(unified.SyncInboundOrderWorkflow)
		w.RegisterWorkflow(unified.InboundOrderPollingWorkflow)

		go func() {
			if err := w.Run(worker.InterruptCh()); err != nil {
				log.Printf("Temporal worker failed: %v", err)
			}
		}()
	}

	// Protected Routes (Require valid JWT)
	r.Group(func(protected chi.Router) {
		protected.Use(auth_middleware.AuthMiddleware(authClient, dbpool))

		organizationHandler.RegisterRoutes(protected)
		auditHandler.RegisterRoutes(protected)
		settingsHandler.RegisterRoutes(protected)
		channelHandler.RegisterRoutes(protected)
		teamWorkspaceHandler.RegisterRoutes(protected)
		pimHandler.RegisterRoutes(protected)
		inventoryHandler.RegisterRoutes(protected)

		// Integrations Callback (Protected)
		protected.Post("/api/v1/integrations/callback", integrationsHandler.HandleSaveConnection)

		// Register Unified Routes
		unifiedHandler.RegisterRoutes(protected, r)

		if err == nil {
			omsHandler := api.NewOMSHandler(temporalClient, queries)
			omsHandler.RegisterRoutes(protected)
		}

		// Register WebSocket endpoints (public & protected)
		websocketHandler.RegisterRoutes(r, protected)

		// Example of safely extracting the injected context claims, protected by RBAC
		protected.With(auth_middleware.RequireRole("admin", "editor")).Get("/api/v1/protected-test", func(w http.ResponseWriter, r *http.Request) {
			tenantID, _ := authcontext.GetTenantID(r.Context())
			orgID, _ := authcontext.GetOrgID(r.Context())
			role, _ := authcontext.GetRole(r.Context())

			w.Write([]byte(fmt.Sprintf("Hello Tenant: %v, Org: %v with Role: %v", tenantID, orgID, role)))
		})
	})

	fmt.Printf("Ops API starting on port %s...\n", port)
	// Note: In production, HTTPS should be handled via a reverse proxy (e.g., Traefik/Cloud Run)
	// or by using http.ListenAndServeTLS if terminating TLS directly in the app.
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func requiredEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("%s is required", name)
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
