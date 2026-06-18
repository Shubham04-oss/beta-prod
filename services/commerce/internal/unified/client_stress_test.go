package unified

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRetryTransport_BodyConsumed(t *testing.T) {
	attemptCount := 0
	var receivedBodies []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++

		bodyBytes, _ := io.ReadAll(r.Body)
		receivedBodies = append(receivedBodies, string(bodyBytes))

		if attemptCount == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("Rate limited"))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL, "test-api-key")

	payload := map[string]interface{}{
		"name": "Test Product",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.PushProduct(ctx, "conn-123", payload)
	if err != nil {
		t.Fatalf("PushProduct failed: %v", err)
	}

	if attemptCount != 2 {
		t.Errorf("Expected 2 attempts, got %d", attemptCount)
	}

	if len(receivedBodies) != 2 {
		t.Fatalf("Expected 2 bodies recorded, got %d", len(receivedBodies))
	}

	if receivedBodies[0] == "" {
		t.Errorf("Attempt 1: body was empty")
	}
	if receivedBodies[1] == "" {
		t.Errorf("Attempt 2: body was empty, meaning retry logic failed to reset body!")
	}
}
