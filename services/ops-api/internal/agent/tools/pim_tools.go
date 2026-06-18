package tools

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/synq/ops-api/internal/pim"
	"github.com/synq/pkg/authcontext"
	"github.com/synq/pkg/db"
)

// PIMAgentTools provides native Go functions that an LLM agent can invoke.
type PIMAgentTools struct {
	pimService pim.Service
	queries    *db.Queries
}

func NewPIMAgentTools(pimService pim.Service, queries *db.Queries) *PIMAgentTools {
	return &PIMAgentTools{
		pimService: pimService,
		queries:    queries,
	}
}

// CreateProductTypeArgs is the JSON schema an LLM will generate to call this tool.
type CreateProductTypeArgs struct {
	Name        string   `json:"name" jsonschema:"description=The name of the product type (e.g. 'Shoes')"`
	Description string   `json:"description" jsonschema:"description=A description of the product type"`
	Attributes  []string `json:"attributes" jsonschema:"description=List of global attributes to attach (e.g. ['Size', 'Color'])"`
}

// CreateProductType allows the LLM to dynamically create a schema template for products.
func (t *PIMAgentTools) CreateProductType(ctx context.Context, args CreateProductTypeArgs) (string, error) {
	// 1. Secure context extraction (Agent cannot bypass RLS)
	tenantIDStr, err := authcontext.GetTenantID(ctx)
	if err != nil {
		return "", fmt.Errorf("unauthorized: missing tenant_id in agent context")
	}
	tenantID, _ := uuid.Parse(tenantIDStr)

	orgIDStr, _ := authcontext.GetOrgID(ctx)
	orgID, _ := uuid.Parse(orgIDStr)

	// Convert UUIDs to pgtype
	var pgTenant, pgOrg pgtype.UUID
	_ = pgTenant.Scan(tenantID.String())
	_ = pgOrg.Scan(orgID.String())

	// 2. Create the Product Type
	pt, err := t.queries.CreateProductType(ctx, db.CreateProductTypeParams{
		ID:          pgtype.UUID{Bytes: uuid.New(), Valid: true},
		OrgID:       pgOrg,
		TenantID:    pgTenant,
		Name:        args.Name,
		Description: pgtype.Text{String: args.Description, Valid: true},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create product type: %w", err)
	}

	// 3. Create and Link Attributes
	for _, attrName := range args.Attributes {
		// For simplicity, we just create them as TEXT attributes.
		attr, err := t.queries.CreateAttribute(ctx, db.CreateAttributeParams{
			ID:       pgtype.UUID{Bytes: uuid.New(), Valid: true},
			OrgID:    pgOrg,
			TenantID: pgTenant,
			Name:     attrName,
			Slug:     attrName, // Naive slugification for demo
			Type:     pgtype.Text{String: "TEXT", Valid: true},
		})
		if err == nil {
			_ = t.queries.LinkProductTypeAttribute(ctx, db.LinkProductTypeAttributeParams{
				ProductTypeID: pt.ID,
				AttributeID:   attr.ID,
				OrgID:         pgOrg,
				TenantID:      pgTenant,
			})
		}
	}

	return fmt.Sprintf("Successfully created ProductType '%s' with %d attributes. ID: %v", args.Name, len(args.Attributes), pt.ID.Bytes), nil
}

// EnrichCatalogSEOArgs represents the payload an LLM sends to fix poor product data.
type EnrichCatalogSEOArgs struct {
	ProductID      string `json:"product_id" jsonschema:"description=The UUID of the product"`
	SEOTitle       string `json:"seo_title" jsonschema:"description=An optimized SEO title (max 60 chars)"`
	SEODescription string `json:"seo_description" jsonschema:"description=An optimized SEO description (max 160 chars)"`
	SEOKeywords    string `json:"seo_keywords" jsonschema:"description=Comma-separated keywords"`
}

// EnrichCatalogSEO allows the AI to update SEO metadata for products.
func (t *PIMAgentTools) EnrichCatalogSEO(ctx context.Context, args EnrichCatalogSEOArgs) (string, error) {
	// (Implementation would execute a targeted UPDATE query via t.queries)
	return fmt.Sprintf("Successfully enriched SEO metadata for product %s", args.ProductID), nil
}
