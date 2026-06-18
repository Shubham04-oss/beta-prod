package unified

import (
	"context"
	"sync"
)

type MockUnifiedClient struct {
	mu      sync.Mutex
	Pushed  map[string]map[string]interface{}
	Orders  map[string]map[string]interface{}
	PushErr error
	PullErr error
	Calls   int
}

func NewMockUnifiedClient() *MockUnifiedClient {
	return &MockUnifiedClient{
		Pushed: make(map[string]map[string]interface{}),
		Orders: make(map[string]map[string]interface{}),
	}
}

func (m *MockUnifiedClient) PushProduct(ctx context.Context, connectionID string, payload map[string]interface{}) (map[string]interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls++

	if m.PushErr != nil {
		return nil, m.PushErr
	}

	id := "mock-id"
	if extID, ok := payload["external_id"].(string); ok && extID != "" {
		id = extID
	}
	m.Pushed[id] = payload

	return map[string]interface{}{
		"id":     id,
		"status": "success",
	}, nil
}

func (m *MockUnifiedClient) PullOrder(ctx context.Context, connectionID string, orderID string) (map[string]interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls++

	if m.PullErr != nil {
		return nil, m.PullErr
	}

	order, ok := m.Orders[orderID]
	if !ok {
		// return a dummy order if not pre-populated
		return map[string]interface{}{
			"id":          orderID,
			"status":      "PAID",
			"currency":    "USD",
			"total_price": 100.0,
		}, nil
	}

	return order, nil
}
