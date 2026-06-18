package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// CommerceClient provides high-speed, connection-pooled access to the internal Commerce API.
type CommerceClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewCommerceClient() *CommerceClient {
	baseURL := os.Getenv("COMMERCE_API_URL")
	if baseURL == "" {
		baseURL = "http://commerce:8080" // Default internal docker-compose routing
	}

	return &CommerceClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

// GetProduct fetches a product from the PIM module
func (c *CommerceClient) GetProduct(ctx context.Context, tenantID, sku string) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/v1/products/%s", c.BaseURL, sku), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Tenant-ID", tenantID)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("commerce API returned status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// CreateOrder sends an order to the OMS module
func (c *CommerceClient) CreateOrder(ctx context.Context, tenantID string, orderPayload interface{}) error {
	body, err := json.Marshal(orderPayload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/v1/orders", c.BaseURL), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("X-Tenant-ID", tenantID)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("commerce API returned status: %d", resp.StatusCode)
	}

	return nil
}
