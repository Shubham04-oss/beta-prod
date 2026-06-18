package telemetry

import (
	"context"
	"testing"
)

func TestSetupTracing_Skip(t *testing.T) {
	ctx := context.Background()
	// An empty endpoint should skip setup but return a valid shutdown function without error
	shutdown, err := SetupTracing(ctx, "test-api", "")
	if err != nil {
		t.Fatalf("Expected no error when skipping telemetry, got: %v", err)
	}
	if shutdown == nil {
		t.Fatalf("Expected valid shutdown func, got nil")
	}

	err = shutdown(ctx)
	if err != nil {
		t.Errorf("Expected nil on dummy shutdown, got: %v", err)
	}
}
