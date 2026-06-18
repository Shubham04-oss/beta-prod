package authcontext

import (
	"context"

	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/workflow"
)

// Context header keys
const (
	HeaderTenantID = "x-tenant-id"
	HeaderOrgID    = "x-org-id"
	HeaderUserID   = "x-user-id"
	HeaderRole     = "x-role"
)

// HeaderWriter abstracts the writer interface used by Temporal (map[string]*commonpb.Payload)
type HeaderWriter interface {
	Set(key string, value *converter.DataConverter) // For custom implementation if needed, but Temporal provides its own maps
}

// PillarContextPropagator implements workflow.ContextPropagator
type PillarContextPropagator struct{}

func NewPillarContextPropagator() workflow.ContextPropagator {
	return &PillarContextPropagator{}
}

// Inject injects values from context into headers for propagation
func (p *PillarContextPropagator) Inject(ctx context.Context, writer workflow.HeaderWriter) error {
	dc := converter.GetDefaultDataConverter()

	if tenantID, err := GetTenantID(ctx); err == nil {
		if payload, err := dc.ToPayload(tenantID); err == nil {
			writer.Set(HeaderTenantID, payload)
		}
	}
	if orgID, err := GetOrgID(ctx); err == nil {
		if payload, err := dc.ToPayload(orgID); err == nil {
			writer.Set(HeaderOrgID, payload)
		}
	}
	if userID, err := GetUserID(ctx); err == nil {
		if payload, err := dc.ToPayload(userID); err == nil {
			writer.Set(HeaderUserID, payload)
		}
	}
	if role, err := GetRole(ctx); err == nil {
		if payload, err := dc.ToPayload(role); err == nil {
			writer.Set(HeaderRole, payload)
		}
	}

	return nil
}

// Extract extracts values from headers and puts them into context
func (p *PillarContextPropagator) Extract(ctx context.Context, reader workflow.HeaderReader) (context.Context, error) {
	dc := converter.GetDefaultDataConverter()

	if payload, ok := reader.Get(HeaderTenantID); ok {
		var tenantID string
		if err := dc.FromPayload(payload, &tenantID); err == nil {
			ctx = WithTenantID(ctx, tenantID)
		}
	}
	if payload, ok := reader.Get(HeaderOrgID); ok {
		var orgID string
		if err := dc.FromPayload(payload, &orgID); err == nil {
			ctx = WithOrgID(ctx, orgID)
		}
	}
	if payload, ok := reader.Get(HeaderUserID); ok {
		var userID string
		if err := dc.FromPayload(payload, &userID); err == nil {
			ctx = WithUserID(ctx, userID)
		}
	}
	if payload, ok := reader.Get(HeaderRole); ok {
		var role string
		if err := dc.FromPayload(payload, &role); err == nil {
			ctx = WithRole(ctx, role)
		}
	}

	return ctx, nil
}

// InjectFromWorkflow injects values from workflow context into headers
func (p *PillarContextPropagator) InjectFromWorkflow(ctx workflow.Context, writer workflow.HeaderWriter) error {
	dc := converter.GetDefaultDataConverter()

	if val := ctx.Value(TenantIDKey); val != nil {
		if payload, err := dc.ToPayload(val.(string)); err == nil {
			writer.Set(HeaderTenantID, payload)
		}
	}
	if val := ctx.Value(OrgIDKey); val != nil {
		if payload, err := dc.ToPayload(val.(string)); err == nil {
			writer.Set(HeaderOrgID, payload)
		}
	}
	if val := ctx.Value(UserIDKey); val != nil {
		if payload, err := dc.ToPayload(val.(string)); err == nil {
			writer.Set(HeaderUserID, payload)
		}
	}
	if val := ctx.Value(RoleKey); val != nil {
		if payload, err := dc.ToPayload(val.(string)); err == nil {
			writer.Set(HeaderRole, payload)
		}
	}

	return nil
}

// ExtractToWorkflow extracts values from headers and puts them into workflow context
func (p *PillarContextPropagator) ExtractToWorkflow(ctx workflow.Context, reader workflow.HeaderReader) (workflow.Context, error) {
	dc := converter.GetDefaultDataConverter()

	if payload, ok := reader.Get(HeaderTenantID); ok {
		var tenantID string
		if err := dc.FromPayload(payload, &tenantID); err == nil {
			ctx = workflow.WithValue(ctx, TenantIDKey, tenantID)
		}
	}
	if payload, ok := reader.Get(HeaderOrgID); ok {
		var orgID string
		if err := dc.FromPayload(payload, &orgID); err == nil {
			ctx = workflow.WithValue(ctx, OrgIDKey, orgID)
		}
	}
	if payload, ok := reader.Get(HeaderUserID); ok {
		var userID string
		if err := dc.FromPayload(payload, &userID); err == nil {
			ctx = workflow.WithValue(ctx, UserIDKey, userID)
		}
	}
	if payload, ok := reader.Get(HeaderRole); ok {
		var role string
		if err := dc.FromPayload(payload, &role); err == nil {
			ctx = workflow.WithValue(ctx, RoleKey, role)
		}
	}

	return ctx, nil
}
