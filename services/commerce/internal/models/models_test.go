package models

import (
	"encoding/json"
	"testing"
)

func TestProductEmbeddingNull(t *testing.T) {
	// Simulate fetching a Product where embedding is NULL.
	// In Go, unmarshaling a JSON null into a *pgvector.Vector should result in nil.
	jsonData := `{"id":"00000000-0000-0000-0000-000000000000", "embedding": null}`
	var p Product
	err := json.Unmarshal([]byte(jsonData), &p)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	if p.Embedding != nil {
		t.Fatalf("Expected Embedding to be nil, got %v", p.Embedding)
	}
}

func TestInventoryLevelDefaults(t *testing.T) {
	// Simulate initializing InventoryLevel without values.
	// Since AvailableQuantity and ReservedQuantity are value types (int), they should default to 0.
	var i InventoryLevel
	if i.AvailableQuantity != 0 {
		t.Fatalf("Expected AvailableQuantity to be 0, got %v", i.AvailableQuantity)
	}
	if i.ReservedQuantity != 0 {
		t.Fatalf("Expected ReservedQuantity to be 0, got %v", i.ReservedQuantity)
	}
}
