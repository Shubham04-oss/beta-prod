package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Mock Unified.to Commerce Item Response
type UnifiedItemResponse struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	Raw       RawData   `json:"raw"`
}

type RawData struct {
	Status string `json:"status"`
	Source string `json:"source"`
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Log the incoming push request
		fmt.Printf("[%s] Received Unified.to %s Request to %s\n", time.Now().Format(time.RFC3339), r.Method, r.URL.Path)

		if r.Method == http.MethodPost {
			// Simulate a successful push to Shopify/WooCommerce
			resp := UnifiedItemResponse{
				ID:        "unified_mock_item_987654321",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Name:      "Mock Synced Product",
				Raw: RawData{
					Status: "success",
					Source: "shopify",
				},
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resp)
			return
		}

		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	fmt.Println("🚀 Unified.to Local Mock Server running on http://localhost:4010")
	log.Fatal(http.ListenAndServe(":4010", nil))
}
