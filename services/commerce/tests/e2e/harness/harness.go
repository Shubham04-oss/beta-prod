package harness

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

// Config holds E2E test configuration.
type Config struct {
	BaseURL string
}

// Harness provides utilities for E2E tests.
type Harness struct {
	Config Config
	Client *http.Client
}

// Setup initializes the test harness.
func Setup(t *testing.T) *Harness {
	t.Helper()
	return &Harness{
		Config: Config{
			BaseURL: "http://localhost:8080/api",
		},
		Client: &http.Client{},
	}
}

// DoRequest performs an HTTP request and unmarshals the JSON response if target is not nil.
func (h *Harness) DoRequest(t *testing.T, method, path string, payload interface{}, target interface{}) int {
	t.Helper()

	var bodyReader io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("Failed to marshal payload: %v", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, h.Config.BaseURL+path, bodyReader)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("X-Org-ID", "00000000-0000-0000-0000-000000000000")
	req.Header.Set("X-Tenant-ID", "00000000-0000-0000-0000-000000000000")
	req.Header.Set("X-Signature", "dummy-signature")

	resp, err := h.Client.Do(req)
	if err != nil {
		// As this is a simulated endpoint, if it's genuinely down, the test will fail.
		// However, standard Go testing requires us to fail the test here.
		// For the sake of test completeness, we will just return 0 or fail.
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if target != nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
		body, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal(body, target); err != nil {
			t.Errorf("Failed to unmarshal response: %v, body: %s", err, string(body))
		}
	}
	b, _ := io.ReadAll(resp.Body)
	t.Logf("Body: %s", string(b))
	return resp.StatusCode
}

// Post is a helper for POST requests.
func (h *Harness) Post(t *testing.T, path string, payload interface{}, target interface{}) int {
	t.Helper()
	return h.DoRequest(t, http.MethodPost, path, payload, target)
}

// Get is a helper for GET requests.
func (h *Harness) Get(t *testing.T, path string, target interface{}) int {
	t.Helper()
	return h.DoRequest(t, http.MethodGet, path, nil, target)
}

// Put is a helper for PUT requests.
func (h *Harness) Put(t *testing.T, path string, payload interface{}, target interface{}) int {
	t.Helper()
	return h.DoRequest(t, http.MethodPut, path, payload, target)
}

// Delete is a helper for DELETE requests.
func (h *Harness) Delete(t *testing.T, path string) int {
	t.Helper()
	return h.DoRequest(t, http.MethodDelete, path, nil, nil)
}

func RandomString(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}
