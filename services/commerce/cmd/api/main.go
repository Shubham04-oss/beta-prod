package main

import (
	"commerce_modules/internal/models"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/jackc/pgx/v5/pgxpool"

	"commerce_modules/internal/db"
	"commerce_modules/internal/inventory"
	"commerce_modules/internal/oms"
	"commerce_modules/internal/pim"
	"commerce_modules/internal/unified"
	"net"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/posthog/posthog-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func getFreePort() uint32 {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer l.Close()
	return uint32(l.Addr().(*net.TCPAddr).Port)
}

func main() {
	// Initialize Sentry
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              os.Getenv("SENTRY_DSN"),
		EnableTracing:    true,
		TracesSampleRate: 1.0,
	})
	if err != nil {
		log.Printf("sentry.Init: %s", err)
	}
	defer sentry.Flush(2 * time.Second)

	// Initialize PostHog
	ph, _ := posthog.NewWithConfig(
		os.Getenv("POSTHOG_API_KEY"),
		posthog.Config{
			Endpoint: "https://app.posthog.com",
		},
	)
	defer ph.Close()

	dbURL := os.Getenv("DATABASE_URL")

	var svc inventory.Service
	var pool *pgxpool.Pool
	var ep *embeddedpostgres.EmbeddedPostgres

	if dbURL != "" {
		pool, err = db.NewDB(context.Background(), dbURL)
		if err != nil {
			log.Fatalf("failed to connect to db: %v", err)
		}
		log.Println("Using pgxpool Postgres DB")
		svc = inventory.NewPgService(pool)
	} else {
		log.Println("DATABASE_URL not set, falling back to embedded postgres")
		tmpDir, _ := os.MkdirTemp("", "main_db_*")
		defer os.RemoveAll(tmpDir)
		port := getFreePort()
		config := embeddedpostgres.DefaultConfig().
			Port(port).
			DataPath(tmpDir + "/data").
			RuntimePath(tmpDir + "/runtime")
		ep = embeddedpostgres.NewDatabase(config)
		if err := ep.Start(); err != nil {
			log.Fatalf("failed to start embedded postgres: %v", err)
		}

		connStr := fmt.Sprintf("postgres://postgres:postgres@localhost:%d/postgres?sslmode=disable", port)
		pool, err = pgxpool.New(context.Background(), connStr)
		if err != nil {
			log.Fatalf("failed to connect to embedded db: %v", err)
		}

		// Execute DDL
		_, err = pool.Exec(context.Background(), `
			CREATE TABLE IF NOT EXISTS locations (
				id UUID PRIMARY KEY,
				org_id UUID NOT NULL,
				tenant_id UUID NOT NULL,
				name TEXT NOT NULL,
				type TEXT NOT NULL,
				metadata JSONB,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				deleted_at TIMESTAMP WITH TIME ZONE
			);
			CREATE TABLE IF NOT EXISTS inventory_levels (
				id UUID PRIMARY KEY,
				org_id UUID NOT NULL,
				tenant_id UUID NOT NULL,
				variant_id UUID NOT NULL,
				location_id UUID NOT NULL,
				available_quantity INTEGER NOT NULL DEFAULT 0,
				reserved_quantity INTEGER NOT NULL DEFAULT 0,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				updated_by UUID,
				UNIQUE(org_id, tenant_id, location_id, variant_id)
			);
			CREATE TABLE IF NOT EXISTS products (
				id UUID PRIMARY KEY,
				org_id UUID NOT NULL,
				tenant_id UUID NOT NULL,
				created_at TIMESTAMP NOT NULL,
				updated_at TIMESTAMP NOT NULL,
				deleted_at TIMESTAMP,
				created_by UUID,
				updated_by UUID,
				title TEXT NOT NULL,
				description TEXT,
				status TEXT NOT NULL,
				options JSONB,
				metadata JSONB,
				embedding text
			);
			CREATE TABLE IF NOT EXISTS product_variants (
				id UUID PRIMARY KEY,
				org_id UUID NOT NULL,
				tenant_id UUID NOT NULL,
				product_id UUID NOT NULL REFERENCES products(id),
				created_at TIMESTAMP NOT NULL,
				updated_at TIMESTAMP NOT NULL,
				deleted_at TIMESTAMP,
				created_by UUID,
				updated_by UUID,
				sku TEXT,
				barcode TEXT,
				currency TEXT NOT NULL,
				price NUMERIC NOT NULL,
				option_values JSONB,
				metadata JSONB,
				UNIQUE(org_id, tenant_id, sku)
			);
			CREATE TABLE IF NOT EXISTS orders (
				id UUID PRIMARY KEY,
				org_id UUID NOT NULL,
				tenant_id UUID NOT NULL,
				customer_id UUID,
				currency VARCHAR(10),
				status VARCHAR(20),
				total_price NUMERIC,
				metadata JSONB,
				created_at TIMESTAMP,
				updated_at TIMESTAMP,
				deleted_at TIMESTAMP
			);
			CREATE TABLE IF NOT EXISTS order_line_items (
				id UUID PRIMARY KEY,
				org_id UUID NOT NULL,
				tenant_id UUID NOT NULL,
				order_id UUID NOT NULL,
				variant_id UUID,
				price_at_purchase NUMERIC,
				option_values_at_purchase JSONB,
				quantity INT,
				created_at TIMESTAMP
			);
			CREATE TABLE IF NOT EXISTS customers (
				id UUID PRIMARY KEY,
				org_id UUID NOT NULL,
				tenant_id UUID NOT NULL,
				first_name VARCHAR(255),
				last_name VARCHAR(255),
				email VARCHAR(255),
				metadata JSONB,
				created_at TIMESTAMP,
				updated_at TIMESTAMP,
				deleted_at TIMESTAMP
			);

		`)
		if err != nil {
			log.Fatalf("failed to execute DDL on embedded db: %v", err)
		}

		svc = inventory.NewPgService(pool)
	}

	defer func() {
		if pool != nil {
			pool.Close()
		}
		if ep != nil {
			if err := ep.Stop(); err != nil {
				log.Printf("failed to stop embedded postgres: %v", err)
			}
		}
	}()

	mux := http.NewServeMux()

	pimRepo := pim.NewRepository()
	pimSvc := pim.NewService(pool, pimRepo, nil) // no inventory cascade for now, or maybe later
	pimAPI := pim.NewAPI(pimSvc)

	omsRepo := oms.NewPostgresRepository(pool)
	omsSvc := oms.NewOMSService(omsRepo, &inventoryAdapter{svc: svc}, &pimAdapter{svc: pimSvc})
	omsAPI := oms.NewAPI(omsSvc)

	inventory.RegisterRoutes(mux, svc)
	pimAPI.RegisterRoutes(mux)
	omsAPI.RegisterRoutes(mux)

	// Prometheus Metrics
	mux.Handle("/metrics", promhttp.Handler())

	var unifiedClient unified.UnifiedClient
	if os.Getenv("UNIFIED_MOCK") == "true" {
		unifiedClient = unified.NewMockUnifiedClient()
	} else {
		unifiedClient = unified.NewHTTPClient(
			os.Getenv("UNIFIED_URL"),
			os.Getenv("UNIFIED_TOKEN"),
		)
	}
	unifiedSvc := unified.NewSyncService(unifiedClient, &pimAdapter{svc: pimSvc}, omsSvc, svc)
	unifiedAPI := unified.NewAPI(unifiedSvc)
	unifiedAPI.RegisterRoutes(mux)

	// TODO: Add service-to-service auth middleware to prevent direct unauthorized access
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		fmt.Println("Server listening on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
	server.Shutdown(context.Background())
}

type inventoryAdapter struct {
	svc inventory.Service
}

func (a *inventoryAdapter) ReserveInventory(ctx context.Context, tenantID, orgID, orderID uuid.UUID, items []models.OrderLineItem) error {
	for i, item := range items {
		if err := a.svc.ReserveStock(ctx, tenantID, orgID, item.VariantID.String(), item.Quantity); err != nil {
			for j := i - 1; j >= 0; j-- {
				a.svc.ReleaseStock(ctx, tenantID, orgID, items[j].VariantID.String(), items[j].Quantity)
			}
			return err
		}
	}
	return nil
}

func (a *inventoryAdapter) DeductInventory(ctx context.Context, tenantID, orgID, orderID uuid.UUID, items []models.OrderLineItem) error {
	for i, item := range items {
		if err := a.svc.ReleaseStock(ctx, tenantID, orgID, item.VariantID.String(), item.Quantity); err != nil {
			for j := i - 1; j >= 0; j-- {
				a.svc.AdjustStock(ctx, tenantID, orgID, items[j].VariantID.String(), items[j].Quantity)
				a.svc.ReserveStock(ctx, tenantID, orgID, items[j].VariantID.String(), items[j].Quantity)
			}
			return err
		}
		if err := a.svc.AdjustStock(ctx, tenantID, orgID, item.VariantID.String(), -item.Quantity); err != nil {
			a.svc.ReserveStock(ctx, tenantID, orgID, item.VariantID.String(), item.Quantity)
			for j := i - 1; j >= 0; j-- {
				a.svc.AdjustStock(ctx, tenantID, orgID, items[j].VariantID.String(), items[j].Quantity)
				a.svc.ReserveStock(ctx, tenantID, orgID, items[j].VariantID.String(), items[j].Quantity)
			}
			return err
		}
	}
	return nil
}

func (a *inventoryAdapter) ReleaseInventory(ctx context.Context, tenantID, orgID, orderID uuid.UUID, items []models.OrderLineItem) error {
	for _, item := range items {
		if err := a.svc.ReleaseStock(ctx, tenantID, orgID, item.VariantID.String(), item.Quantity); err != nil {
			return err
		}
	}
	return nil
}

type pimAdapter struct {
	svc *pim.Service
}

func (a *pimAdapter) GetVariant(ctx context.Context, tenantID, orgID, variantID uuid.UUID) (*models.ProductVariant, error) {
	variant, err := a.svc.GetVariant(ctx, orgID, tenantID, variantID)
	if err == nil && variant != nil {
		return variant, nil
	}

	// Fallback: variantID might actually be a productID in our E2E tests and legacy endpoints
	variants, err := a.svc.ListVariants(ctx, orgID, tenantID, variantID)
	fmt.Printf("DEBUG GetVariant Fallback: orgID=%s, tenantID=%s, variantID(as productID)=%s, found=%d, err=%v\n", orgID, tenantID, variantID, len(variants), err)
	if err != nil || len(variants) == 0 {
		return nil, fmt.Errorf("variant not found for id %s", variantID)
	}
	return variants[0], nil
}

func (a *pimAdapter) GetProduct(ctx context.Context, tenantID, orgID, productID uuid.UUID) (*models.Product, error) {
	return a.svc.GetProduct(ctx, orgID, tenantID, productID)
}

func (a *pimAdapter) ListVariants(ctx context.Context, tenantID, orgID, productID uuid.UUID) ([]*models.ProductVariant, error) {
	return a.svc.ListVariants(ctx, orgID, tenantID, productID)
}
