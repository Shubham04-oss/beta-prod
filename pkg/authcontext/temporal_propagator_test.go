package authcontext_test

import (
	"context"
	"testing"

	"github.com/synq/pkg/authcontext"
	"go.temporal.io/api/common/v1"
)

// MockHeader implements both workflow.HeaderWriter and workflow.HeaderReader
type MockHeader struct {
	Payloads map[string]*common.Payload
}

func NewMockHeader() *MockHeader {
	return &MockHeader{
		Payloads: make(map[string]*common.Payload),
	}
}

func (m *MockHeader) Set(key string, value *common.Payload) {
	m.Payloads[key] = value
}

func (m *MockHeader) Get(key string) (*common.Payload, bool) {
	val, ok := m.Payloads[key]
	return val, ok
}

func (m *MockHeader) ForEachKey(handler func(key string, payload *common.Payload) error) error {
	for k, v := range m.Payloads {
		if err := handler(k, v); err != nil {
			return err
		}
	}
	return nil
}

func TestTemporalPropagator_Context(t *testing.T) {
	propagator := authcontext.NewPillarContextPropagator()

	// 1. Create a base context and inject 4 pillars
	ctx := context.Background()
	ctx = authcontext.WithTenantID(ctx, "tenant-123")
	ctx = authcontext.WithOrgID(ctx, "org-456")
	ctx = authcontext.WithUserID(ctx, "user-789")
	ctx = authcontext.WithRole(ctx, "admin")

	// 2. Inject into Mock Header (Simulating sending over network)
	header := NewMockHeader()
	err := propagator.Inject(ctx, header)
	if err != nil {
		t.Fatalf("Failed to inject context into header: %v", err)
	}

	// 3. Extract into a new empty context (Simulating receiving on worker)
	emptyCtx := context.Background()
	extractedCtx, err := propagator.Extract(emptyCtx, header)
	if err != nil {
		t.Fatalf("Failed to extract context from header: %v", err)
	}

	// 4. Verify 4 pillars survived the round trip
	tenantID, err := authcontext.GetTenantID(extractedCtx)
	if err != nil || tenantID != "tenant-123" {
		t.Errorf("Expected TenantID 'tenant-123', got '%s' (err: %v)", tenantID, err)
	}

	orgID, err := authcontext.GetOrgID(extractedCtx)
	if err != nil || orgID != "org-456" {
		t.Errorf("Expected OrgID 'org-456', got '%s' (err: %v)", orgID, err)
	}

	userID, err := authcontext.GetUserID(extractedCtx)
	if err != nil || userID != "user-789" {
		t.Errorf("Expected UserID 'user-789', got '%s' (err: %v)", userID, err)
	}

	role, err := authcontext.GetRole(extractedCtx)
	if err != nil || role != "admin" {
		t.Errorf("Expected Role 'admin', got '%s' (err: %v)", role, err)
	}
}
