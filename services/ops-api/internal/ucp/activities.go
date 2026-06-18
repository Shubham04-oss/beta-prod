package ucp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/synq/pkg/db"
	"github.com/synq/ops-api/internal/telemetry"
)

type Activities struct {
	pool    *pgxpool.Pool
	storage *storage.Client
}

func NewActivities(pool *pgxpool.Pool, sc *storage.Client) *Activities {
	return &Activities{
		pool:    pool,
		storage: sc,
	}
}

// ExtractCatalogActivity pulls the PIM data and maps it to UCP structs
func (a *Activities) ExtractCatalogActivity(ctx context.Context, tenantIDStr string) (UCPFeed, error) {
	tenantUUID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	queries := db.New(a.pool)
	var dbTenantID pgtype.UUID
	err = dbTenantID.Scan(tenantUUID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to scan tenant UUID: %w", err)
	}

	rows, err := queries.ListProductsForUCP(ctx, dbTenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list products for UCP: %w", err)
	}

	// Group rows into products
	productMap := make(map[string]*UCPProduct)

	for _, row := range rows {
		prodID := uuidFromPgType(row.ProductID)
		
		prod, exists := productMap[prodID]
		if !exists {
			prod = &UCPProduct{
				Context:     "https://schema.org/",
				Type:        "ProductGroup",
				ProductID:   prodID,
				Name:        row.Title,
				Description: stringFromPgType(row.Description),
				Category:    stringFromPgType(row.Category),
				HasVariant:  []UCPProductVariant{},
			}
			brandStr := stringFromPgType(row.Brand)
			if brandStr != "" {
				prod.Brand = &UCPBrand{
					Type: "Brand",
					Name: brandStr,
				}
			}
			productMap[prodID] = prod
		}

		imgUrl := stringFromPgType(row.ImageUrl)
		if imgUrl != "" {
			// Basic deduplication
			foundImage := false
			for _, img := range prod.Image {
				if img == imgUrl {
					foundImage = true
					break
				}
			}
			if !foundImage {
				prod.Image = append(prod.Image, imgUrl)
			}
		}

		if row.VariantID.Valid {
			var priceStr string
			if row.Price.Valid {
				if row.Price.Int != nil {
					priceStr = fmt.Sprintf("%.2f", float64(row.Price.Int.Int64())/100.0)
				}
			}

			variant := UCPProductVariant{
				Type: "Product",
				SKU:  stringFromPgType(row.Sku),
				GTIN: stringFromPgType(row.Gtin),
				Name: row.Title, // Inherits base name
				Offers: &UCPOffer{
					Type:          "Offer",
					Price:         priceStr,
					PriceCurrency: stringFromPgType(row.Currency),
					Availability:  "https://schema.org/InStock",
				},
			}
			prod.HasVariant = append(prod.HasVariant, variant)
		}
	}

	var feed UCPFeed
	for _, p := range productMap {
		feed = append(feed, *p)
	}

	return feed, nil
}

// UploadFeedToGCSActivity uploads the JSON-LD string to GCS
func (a *Activities) UploadFeedToGCSActivity(ctx context.Context, feed UCPFeed, tenantID string) (string, error) {
	bucketName := "ucp-feeds"
	objectName := fmt.Sprintf("tenant_%s_ucp_catalog.json", tenantID)

	jsonData, err := json.MarshalIndent(feed, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal UCP feed: %w", err)
	}

	bucket := a.storage.Bucket(bucketName)
	
	// Create bucket if it doesn't exist (helpful for local emulator)
	if _, err := bucket.Attrs(ctx); err != nil {
		if err := bucket.Create(ctx, "demo-synq", nil); err != nil {
			log.Printf("Bucket creation warning (may already exist): %v", err)
		}
	}

	obj := bucket.Object(objectName)
	writer := obj.NewWriter(ctx)
	writer.ContentType = "application/json"

	if _, err := writer.Write(jsonData); err != nil {
		return "", fmt.Errorf("failed to write object data: %w", err)
	}
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close object writer: %w", err)
	}

	// Assuming local emulator or public access
	url := fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, objectName)
	log.Printf("Successfully uploaded UCP feed to %s", url)

	// Record Prometheus Metric
	telemetry.UCPFeedGenerationDuration.WithLabelValues(tenantID, "success").Observe(1.0) // simplified observation for now

	return url, nil
}

func stringFromPgType(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}

func uuidFromPgType(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	parsed, _ := uuid.FromBytes(u.Bytes[:])
	return parsed.String()
}
