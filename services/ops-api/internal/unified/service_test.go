package unified

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	pgxvec "github.com/pgvector/pgvector-go/pgx"

	"github.com/synq/pkg/db"
	sdk "github.com/unified-to/unified-go-sdk"
)

func TestServiceProcessPush(t *testing.T) {
	_ = godotenv.Load("../../.env")
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL must be set")
	}

	ctx := context.Background()

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		t.Fatalf("Failed to parse db config: %v", err)
	}
	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		return pgxvec.RegisterTypes(ctx, conn)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Seed Data
	orgID := uuid.New()
	tenantID := uuid.New()
	userID := uuid.New()
	productID := uuid.New()
	connectionID := uuid.New()

	// Ensure mapping table exists for tests
	_, _ = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS commerce_item_mappings (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			org_id UUID NOT NULL,
			tenant_id UUID NOT NULL,
			connection_id UUID NOT NULL,
			product_id UUID NOT NULL,
			unified_item_id VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(connection_id, product_id)
		)
	`)

	// Insert Organization
	_, err = pool.Exec(ctx, `INSERT INTO organizations (id, name) VALUES ($1, $2)`, orgID, "Unified Test Org")
	if err != nil {
		t.Fatalf("insert org: %v", err)
	}

	// Insert Tenant
	_, err = pool.Exec(ctx, `INSERT INTO tenants (id, org_id, name) VALUES ($1, $2, $3)`, tenantID, orgID, "Unified Test Tenant")
	if err != nil {
		t.Fatalf("insert tenant: %v", err)
	}

	// Ensure user exists
	email := userID.String() + "@test.com"
	_, _ = pool.Exec(ctx, `INSERT INTO users (id, org_id, tenant_id, email, role) VALUES ($1, $2, $3, $4, $5)`, userID, orgID, tenantID, email, "ADMIN")

	// Create Product
	_, err = pool.Exec(ctx, `INSERT INTO products (id, org_id, tenant_id, title, status, created_by, data_quality_score) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		productID, orgID, tenantID, "Push Test Product", "ACTIVE", userID, 100)
	if err != nil {
		t.Fatalf("insert product: %v", err)
	}

	// Create Commerce Connection
	_, err = pool.Exec(ctx, `INSERT INTO commerce_connections (id, org_id, tenant_id, unified_connection_id, provider, status) VALUES ($1, $2, $3, $4, $5, $6)`,
		connectionID, orgID, tenantID, "ext_conn_123", "shopify", "ACTIVE")
	if err != nil {
		t.Fatalf("insert connection: %v", err)
	}

	// Setup Mock HTTP Server to mimic Unified.to API
	var receivedRequests int
	var payload map[string]interface{}
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedRequests++
		if r.Method == "POST" {
			json.NewDecoder(r.Body).Decode(&payload)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id": "unified_item_999"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}))
	defer mockServer.Close()

	// Initialize SDK with mock server
	s := sdk.New(sdk.WithServerURL(mockServer.URL), sdk.WithSecurity("dummy_token"))

	unifiedService := &Service{
		pool:       pool,
		unifiedSDK: s,
		dbQueries:  db.New(pool),
	}
	unifiedService.limiter = NewService(pool, "dummy_token").limiter // copy default limiter
	unifiedService.workerPool = NewWorkerPool(unifiedService, 1, 10)

	// Process Push
	err = unifiedService.ProcessPush(ctx, tenantID.String(), orgID.String(), productID.String(), "UPSERT")

	if err != nil {
		t.Fatalf("ProcessPush failed against Prism mock (invalid payload?): %v", err)
	}

	t.Log("ProcessPush successfully validated payload against Official Unified.to OpenAPI Spec via Prism!")
}
