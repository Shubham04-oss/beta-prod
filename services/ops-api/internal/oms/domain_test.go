package oms_test

import (
	"testing"

	"github.com/synq/ops-api/internal/oms"
)

func TestOrderStateTransitions(t *testing.T) {
	tests := []struct {
		name        string
		current     oms.OrderStatus
		target      oms.OrderStatus
		shouldAllow bool
	}{
		{"Draft to PendingPayment", oms.StatusDraft, oms.StatusPendingPayment, true},
		{"PendingPayment to PaymentAuthorized", oms.StatusPendingPayment, oms.StatusPaymentAuthorized, true},
		{"PaymentAuthorized to Confirmed", oms.StatusPaymentAuthorized, oms.StatusConfirmed, true},
		{"Confirmed to Processing", oms.StatusConfirmed, oms.StatusProcessing, true},
		{"Processing to PartiallyFulfilled", oms.StatusProcessing, oms.StatusPartiallyFulfilled, true},
		{"PartiallyFulfilled to Fulfilled", oms.StatusPartiallyFulfilled, oms.StatusFulfilled, true},
		{"Processing to Fulfilled", oms.StatusProcessing, oms.StatusFulfilled, true},
		{"PendingPayment to Cancelled", oms.StatusPendingPayment, oms.StatusCancelled, true},
		{"Fulfilled to ReturnRequested", oms.StatusFulfilled, oms.StatusReturnRequested, true},
		
		// Invalid transitions
		{"Draft to Fulfilled", oms.StatusDraft, oms.StatusFulfilled, false},
		{"Cancelled to Confirmed", oms.StatusCancelled, oms.StatusConfirmed, false},
		{"Fulfilled to Draft", oms.StatusFulfilled, oms.StatusDraft, false},
		{"Draft to Confirmed", oms.StatusDraft, oms.StatusConfirmed, false}, // Must go through payment states
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := &oms.Order{Status: tt.current}
			allowed := order.CanTransitionTo(tt.target)
			if tt.shouldAllow && !allowed {
				t.Errorf("Expected transition %s -> %s to be allowed, but it was denied", tt.current, tt.target)
			}
			if !tt.shouldAllow && allowed {
				t.Errorf("Expected transition %s -> %s to be denied, but it was allowed", tt.current, tt.target)
			}
		})
	}
}

func TestOrderLineItemCalculations(t *testing.T) {
	// Example test placeholder for line item tax/discount/total calculations if added later
}
