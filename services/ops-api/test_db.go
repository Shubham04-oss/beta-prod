package main

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/pubsub"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	pgxvec "github.com/pgvector/pgvector-go/pgx"

	"github.com/synq/ops-api/internal/pim"
	"github.com/synq/pkg/authcontext"
	"github.com/synq/pkg/db"
	"github.com/synq/pkg/events"
)

func createTopicIfNotExists(ctx context.Context, projectID, topicID string) error {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return err
	}
	defer client.Close()

	topic := client.Topic(topicID)
	exists, err := topic.Exists(ctx)
	if err != nil {
		return err
	}
	if !exists {
		log.Printf("Creating topic %s", topicID)
		_, err = client.CreateTopic(ctx, topicID)
		return err
	}
	return nil
}

func runFile(ctx context.Context, pool *pgxpool.Pool, filepath string) error {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, string(content))
	if err != nil {
		log.Printf("Warning executing %s: %v", filepath, err)
	} else {
		log.Printf("Successfully executed %s", filepath)
	}
	return nil
}

func main() {
	_ = godotenv.Load(".env")
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL must be set")
	}

	ctx := context.Background()

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		log.Fatalf("Failed to parse db config: %v", err)
	}
	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		return pgxvec.RegisterTypes(ctx, conn)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	log.Println("Connected to database successfully")

	// 1. Run Migrations
	_ = runFile(ctx, pool, "../../infrastructure/postgres/init.sql")
	_ = runFile(ctx, pool, "../../infrastructure/postgres/pim.sql")
	_ = runFile(ctx, pool, "../../infrastructure/postgres/pim_enrichment.sql")

	// 2. Setup Seed Data (Bypassing RLS by using standard queries)

	orgID := uuid.New()
	tenantID := uuid.New()
	userID := uuid.New()

	// Insert Organization
	_, err = pool.Exec(ctx, `INSERT INTO organizations (id, name) VALUES ($1, $2) ON CONFLICT DO NOTHING`, orgID, "Test Org")
	// Insert Tenant
	_, err = pool.Exec(ctx, `INSERT INTO tenants (id, org_id, name) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`, tenantID, orgID, "Test Tenant")
	// Insert User
	email := userID.String() + "@test.com"
	_, err = pool.Exec(ctx, `INSERT INTO users (id, org_id, tenant_id, email, role) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (email) DO NOTHING`, userID, orgID, tenantID, email, "ADMIN")

	log.Printf("Seeded Organization: %s", orgID)
	log.Printf("Seeded Tenant: %s", tenantID)

	// 3. Create context with the 4 pillars
	reqCtx := context.WithValue(ctx, authcontext.TenantIDKey, tenantID.String())
	reqCtx = context.WithValue(reqCtx, authcontext.OrgIDKey, orgID.String())
	reqCtx = context.WithValue(reqCtx, authcontext.UserIDKey, userID.String())
	reqCtx = context.WithValue(reqCtx, authcontext.RoleKey, "ADMIN")

	// 4. Instantiate PIM Service
	// Initialize PubSub Publisher
	if err := createTopicIfNotExists(ctx, "demo-synq", "pim-events"); err != nil {
		log.Fatalf("Failed to create topic: %v", err)
	}

	publisher, err := events.NewPubSubPublisher(ctx, "demo-synq")
	if err != nil {
		log.Fatalf("Failed to initialize pubsub publisher: %v", err)
	}
	defer publisher.Close()

	// Initialize PIM Service
	pimService := pim.NewService(pool, publisher)

	// 5. Test Create Product
	log.Println("Testing CreateProduct...")
	product, err := pimService.CreateProduct(reqCtx, db.CreateProductParams{
		ID:               pgtype.UUID{Bytes: uuid.New(), Valid: true},
		OrgID:            pgtype.UUID{Bytes: orgID, Valid: true},
		TenantID:         pgtype.UUID{Bytes: tenantID, Valid: true},
		Title:            "E2E Test Headphones",
		ShortDescription: pgtype.Text{String: "Great headphones", Valid: true},
		SeoTitle:         pgtype.Text{String: "Buy E2E Test Headphones", Valid: true},
		Category:         pgtype.Text{String: "Audio", Valid: true},
		CreatedBy:        pgtype.UUID{Bytes: userID, Valid: true},
	})
	if err != nil {
		log.Fatalf("Failed to create product: %v", err)
	}
	log.Printf("Success! Created Product: %s | Title: %s", product.ID.Bytes, product.Title)

	// 6. Test Create Variant
	log.Println("Testing CreateVariant...")
	variant, err := pimService.CreateVariant(reqCtx, db.CreateProductVariantParams{
		ID:        pgtype.UUID{Bytes: uuid.New(), Valid: true},
		OrgID:     pgtype.UUID{Bytes: orgID, Valid: true},
		TenantID:  pgtype.UUID{Bytes: tenantID, Valid: true},
		ProductID: product.ID,
		Sku:       pgtype.Text{String: "E2E-TEST-001", Valid: true},
		Price:     pgtype.Numeric{Int: nil, Valid: true}, // Mocking 0 price
	})
	if err != nil {
		log.Printf("Failed to create variant: %v", err)
	} else {
		log.Printf("Success! Created Variant: %s | SKU: %s", variant.ID.Bytes, variant.Sku.String)
	}

	// 7. Test Create Media
	log.Println("Testing CreateMedia...")
	media, err := pimService.CreateMedia(reqCtx, db.CreateProductMediaParams{
		ID:        pgtype.UUID{Bytes: uuid.New(), Valid: true},
		OrgID:     pgtype.UUID{Bytes: orgID, Valid: true},
		TenantID:  pgtype.UUID{Bytes: tenantID, Valid: true},
		ProductID: product.ID,
		Url:       "https://cdn.example.com/headphones.png",
		AltText:   pgtype.Text{String: "Front view of headphones", Valid: true},
		SortOrder: pgtype.Int4{Int32: 1, Valid: true},
	})
	if err != nil {
		log.Fatalf("Failed to create media: %v", err)
	}
	log.Printf("Success! Created Media: %s | URL: %s", media.ID.Bytes, media.Url)

	log.Println("PIM Backend E2E Test Completed Successfully!")
}
