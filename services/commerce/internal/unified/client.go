package unified

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type UnifiedClient interface {
	PushProduct(ctx context.Context, connectionID string, payload map[string]interface{}) (map[string]interface{}, error)
	PullOrder(ctx context.Context, connectionID string, orderID string) (map[string]interface{}, error)
}

type httpUnifiedClient struct {
	client  *http.Client
	baseURL string
	apiKey  string
}

func NewHTTPClient(baseURL, apiKey string) UnifiedClient {
	return &httpUnifiedClient{
		client: &http.Client{
			Transport: &retryTransport{
				base: http.DefaultTransport,
			},
			Timeout: 30 * time.Second,
		},
		baseURL: baseURL,
		apiKey:  apiKey,
	}
}

func (c *httpUnifiedClient) PushProduct(ctx context.Context, connectionID string, payload map[string]interface{}) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/commerce/%s/product", c.baseURL, connectionID)

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *httpUnifiedClient) PullOrder(ctx context.Context, connectionID string, orderID string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/commerce/%s/order/%s", c.baseURL, connectionID, orderID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// retryTransport implements http.RoundTripper with exponential backoff on 429 and 5xx
type retryTransport struct {
	base http.RoundTripper
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	maxRetries := 3
	backoff := 100 * time.Millisecond

	for i := 0; i <= maxRetries; i++ {
		if i > 0 && req.GetBody != nil {
			body, bodyErr := req.GetBody()
			if bodyErr != nil {
				return nil, bodyErr
			}
			req.Body = body
		}

		resp, err = t.base.RoundTrip(req)

		if err == nil {
			if resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode < 500 {
				return resp, nil
			}

			// Needs retry but is last attempt
			if i == maxRetries {
				return resp, nil
			}

			// Close body before retrying
			resp.Body.Close()
		} else {
			// Request error
			if i == maxRetries {
				return nil, err
			}
		}

		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		case <-time.After(backoff):
		}
		backoff *= 2
	}

	return resp, err
}
