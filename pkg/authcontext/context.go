package authcontext

import (
	"context"
	"fmt"
)

type ContextKey string

const (
	TenantIDKey ContextKey = "tenant_id"
	OrgIDKey    ContextKey = "org_id"
	UserIDKey   ContextKey = "user_id"
	RoleKey     ContextKey = "role"
)

// Inject context fields
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, TenantIDKey, tenantID)
}

func WithOrgID(ctx context.Context, orgID string) context.Context {
	return context.WithValue(ctx, OrgIDKey, orgID)
}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

func WithRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, RoleKey, role)
}

// Extract context fields securely
func GetTenantID(ctx context.Context) (string, error) {
	val := ctx.Value(TenantIDKey)
	if val == nil {
		return "", fmt.Errorf("authcontext: missing tenant_id")
	}
	str, ok := val.(string)
	if !ok || str == "" {
		return "", fmt.Errorf("authcontext: invalid tenant_id type or empty")
	}
	return str, nil
}

func GetOrgID(ctx context.Context) (string, error) {
	val := ctx.Value(OrgIDKey)
	if val == nil {
		return "", fmt.Errorf("authcontext: missing org_id")
	}
	str, ok := val.(string)
	if !ok || str == "" {
		return "", fmt.Errorf("authcontext: invalid org_id type or empty")
	}
	return str, nil
}

func GetUserID(ctx context.Context) (string, error) {
	val := ctx.Value(UserIDKey)
	if val == nil {
		return "", fmt.Errorf("authcontext: missing user_id")
	}
	str, ok := val.(string)
	if !ok || str == "" {
		return "", fmt.Errorf("authcontext: invalid user_id type or empty")
	}
	return str, nil
}

func GetRole(ctx context.Context) (string, error) {
	val := ctx.Value(RoleKey)
	if val == nil {
		return "", fmt.Errorf("authcontext: missing role")
	}
	str, ok := val.(string)
	if !ok || str == "" {
		return "", fmt.Errorf("authcontext: invalid role type or empty")
	}
	return str, nil
}
