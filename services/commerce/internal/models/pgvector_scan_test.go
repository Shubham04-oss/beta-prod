package models_test

import (
	"commerce_modules/internal/models"
	"database/sql"
	"testing"

	"github.com/pgvector/pgvector-go"
)

// We can test database/sql conversion by implementing a simple mock driver or just using sql.Scanner interface.
func TestPgVectorScanInterface(t *testing.T) {
	var p models.Product

	// Ensure that *pgvector.Vector implements sql.Scanner
	var _ sql.Scanner = (*pgvector.Vector)(nil)

	// If p.Embedding is *pgvector.Vector, when we pass &p.Embedding to rows.Scan,
	// database/sql uses reflection. Let's make sure pointer type is correct.
	if p.Embedding != nil {
		t.Fatalf("expected nil")
	}
}
