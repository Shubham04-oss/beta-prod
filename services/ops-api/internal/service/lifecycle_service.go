package service

import (
	"context"
	"fmt"

	"firebase.google.com/go/v4/auth"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LifecycleService struct {
	db   *pgxpool.Pool
	auth *auth.Client
}

func NewLifecycleService(db *pgxpool.Pool, authClient *auth.Client) *LifecycleService {
	return &LifecycleService{
		db:   db,
		auth: authClient,
	}
}

// CreateTenantLifecycle handles creating an Organization, a Tenant, and the initial Admin User
func (s *LifecycleService) CreateTenantLifecycle(ctx context.Context, orgName, tenantName, adminEmail, adminPassword string) error {
	// 1. Database Transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Create Organization
	var orgID string
	err = tx.QueryRow(ctx, "INSERT INTO organizations (name) VALUES ($1) RETURNING id", orgName).Scan(&orgID)
	if err != nil {
		return fmt.Errorf("failed to insert org: %w", err)
	}

	// Create Tenant
	var tenantID string
	err = tx.QueryRow(ctx, "INSERT INTO tenants (org_id, name) VALUES ($1, $2) RETURNING id", orgID, tenantName).Scan(&tenantID)
	if err != nil {
		return fmt.Errorf("failed to insert tenant: %w", err)
	}

	// 2. We skip Firebase Identity Platform Multi-Tenancy to allow global login from the frontend.
	// We rely purely on Postgres and Custom Claims for strict multi-tenant isolation.
	// 3. Create Admin User in Firebase globally
	params := (&auth.UserToCreate{}).
		Email(adminEmail).
		Password(adminPassword).
		DisplayName("Admin")

	fbUser, err := s.auth.CreateUser(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to create firebase user: %w", err)
	}

	// 4. Create User in Database
	var userID string
	err = tx.QueryRow(ctx,
		"INSERT INTO users (org_id, tenant_id, email, role) VALUES ($1, $2, $3, 'ADMIN') RETURNING id",
		orgID, tenantID, adminEmail).Scan(&userID)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	// 5. Assign Custom Claims to Firebase User
	claims := map[string]interface{}{
		"org_id":    orgID,
		"tenant_id": tenantID,
		"role":      "ADMIN",
		"db_uid":    userID,
	}
	if err := s.auth.SetCustomUserClaims(ctx, fbUser.UID, claims); err != nil {
		return fmt.Errorf("failed to set custom claims: %w", err)
	}

	// Commit Transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// InviteUser handles provisioning a new team member directly into an existing Organization/Tenant.
func (s *LifecycleService) InviteUser(ctx context.Context, email, orgID, tenantID, role string) (*auth.UserRecord, error) {
	// 1. Database Transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 2. Create User in Firebase
	params := (&auth.UserToCreate{}).
		Email(email).
		EmailVerified(false)

	fbUser, err := s.auth.CreateUser(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create firebase user: %w", err)
	}

	// 3. Create User in Database
	var userID string
	err = tx.QueryRow(ctx,
		"INSERT INTO users (org_id, tenant_id, email, role) VALUES ($1, $2, $3, $4) RETURNING id",
		orgID, tenantID, email, role).Scan(&userID)
	if err != nil {
		// Clean up firebase user if DB insert fails
		_ = s.auth.DeleteUser(ctx, fbUser.UID)
		return nil, fmt.Errorf("failed to insert user into db: %w", err)
	}

	// 4. Assign Custom Claims
	claims := map[string]interface{}{
		"org_id":    orgID,
		"tenant_id": tenantID,
		"role":      role,
		"db_uid":    userID,
	}
	if err := s.auth.SetCustomUserClaims(ctx, fbUser.UID, claims); err != nil {
		return nil, fmt.Errorf("failed to set custom claims: %w", err)
	}

	// 5. Generate Password Reset Link (sent to user so they can set initial password)
	// In emulator, this prints to the console.
	_, err = s.auth.PasswordResetLink(ctx, email)
	if err != nil {
		// Non-fatal, just log it
		fmt.Printf("Warning: failed to generate password reset link for %s: %v\n", email, err)
	}

	// Commit Transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return fbUser, nil
}
