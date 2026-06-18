package pim

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/synq/pkg/authcontext"
	"github.com/synq/pkg/db"
)

// MockPublisher is a mock of the events.Publisher interface
type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) Publish(ctx context.Context, topicID, eventType string, payload interface{}) error {
	args := m.Called(ctx, topicID, eventType, payload)
	return args.Error(0)
}

func (m *MockPublisher) Close() error {
	args := m.Called()
	return args.Error(0)
}

// TestContextWithPillars creates a mock context containing the 4 IDs
func TestContextWithPillars() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, authcontext.TenantIDKey, "tenant-123")
	ctx = context.WithValue(ctx, authcontext.OrgIDKey, "org-456")
	ctx = context.WithValue(ctx, authcontext.UserIDKey, "user-789")
	ctx = context.WithValue(ctx, authcontext.RoleKey, "admin")
	return ctx
}

func TestEnforceRLS_ContextExtraction(t *testing.T) {
	// We want to ensure that the service correctly extracts the TenantID from context
	ctx := TestContextWithPillars()
	
	tenantID, err := authcontext.GetTenantID(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "tenant-123", tenantID)

	orgID, err := authcontext.GetOrgID(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "org-456", orgID)
}

// NOTE: A full integration test would require a live Postgres container with pgxpool.
// This test file outlines the verification structure:
// 1. Initialize a pgxpool connected to the test DB.
// 2. Pass in MockPublisher.
// 3. Call CreateProduct() with the TestContextWithPillars.
// 4. Verify that MockPublisher.AssertCalled(t, "Publish", ctx, "pim-events", "synq.pim.product.created", mock.Anything) is successful.
// 5. Verify the db.Product returned has the correct TenantID from the RLS enforcement.
