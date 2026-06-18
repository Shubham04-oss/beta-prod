package tools

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/synq/pkg/authcontext"
	"github.com/synq/pkg/db"
	"github.com/jackc/pgx/v5/pgtype"
	unifiedgosdk "github.com/unified-to/unified-go-sdk"
	"github.com/unified-to/unified-go-sdk/pkg/models/operations"
)

type UnifiedActivities struct {
	queries *db.Queries
	unified *unifiedgosdk.UnifiedTo
}

func NewUnifiedActivities(dbpool *pgxpool.Pool, workspaceID string) *UnifiedActivities {
	u := unifiedgosdk.New(
		unifiedgosdk.WithSecurity(workspaceID),
	)

	return &UnifiedActivities{
		queries: db.New(dbpool),
		unified: u,
	}
}

// FetchCommerceItems is a Temporal Activity designed for LLM Agents to consume.
// It securely retrieves the tenant's connection_id using sqlc and fetches their items via Unified.to.
func (a *UnifiedActivities) FetchCommerceItems(ctx context.Context) (string, error) {
	// 1. Extract the secure tenant context
	tenantIDStr, err := authcontext.GetTenantID(ctx)
	if err != nil {
		return "", fmt.Errorf("security violation: %v", err)
	}

	var tenantUUID pgtype.UUID
	err = tenantUUID.Scan(tenantIDStr)
	if err != nil {
		return "", fmt.Errorf("invalid tenant UUID format: %v", err)
	}

	// 2. Fetch the Connection ID from Postgres via sqlc
	connectionID, err := a.queries.GetTenantIntegrationByCategory(ctx, db.GetTenantIntegrationByCategoryParams{
		TenantID: tenantUUID,
		Category: db.IntegrationCategoryCommerce,
	})
	
	if err != nil {
		return "", fmt.Errorf("no commerce integration found for tenant %s: %v", tenantIDStr, err)
	}

	log.Printf("Found connection %s for tenant %s. Fetching items...", connectionID, tenantIDStr)

	// 2. Fetch Commerce Items from Unified.to API
	req := operations.ListCommerceItemsRequest{
		ConnectionID: connectionID,
	}

	res, err := a.unified.Commerce.ListCommerceItems(ctx, req)
	if err != nil {
		return "", fmt.Errorf("unified.to API error: %v", err)
	}

	// 3. Format the result for the LLM
	if len(res.CommerceItems) == 0 {
		return "No commerce items found.", nil
	}

	summary := fmt.Sprintf("Found %d items:\n", len(res.CommerceItems))
	for _, item := range res.CommerceItems {
		name := "Unknown"
		if item.Name != nil {
			name = *item.Name
		}
		
		// Safely handle missing IDs
		id := "Unknown"
		if item.ID != nil {
			id = *item.ID
		}
		
		summary += fmt.Sprintf("- ID: %s, Name: %s\n", id, name)
	}

	return summary, nil
}
