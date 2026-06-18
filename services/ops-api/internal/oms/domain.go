package oms

import (
	"fmt"
	"time"
)

// OrderStatus matches the order_status ENUM in Postgres
type OrderStatus string

const (
	StatusDraft              OrderStatus = "draft"
	StatusPendingPayment     OrderStatus = "pending_payment"
	StatusPaymentAuthorized  OrderStatus = "payment_authorized"
	StatusConfirmed          OrderStatus = "confirmed"
	StatusProcessing         OrderStatus = "processing"
	StatusPartiallyFulfilled OrderStatus = "partially_fulfilled"
	StatusFulfilled          OrderStatus = "fulfilled"
	StatusDelivered          OrderStatus = "delivered"
	StatusCompleted          OrderStatus = "completed"
	StatusCancelled          OrderStatus = "cancelled"
	StatusReturnRequested    OrderStatus = "return_requested"
	StatusPartiallyReturned  OrderStatus = "partially_returned"
	StatusReturned           OrderStatus = "returned"
	StatusRefunded           OrderStatus = "refunded"
	StatusPartiallyRefunded  OrderStatus = "partially_refunded"
	StatusFailed             OrderStatus = "failed"
)

// FulfillmentStatus matches the fulfillment_status ENUM in Postgres
type FulfillmentStatus string

const (
	FulfillmentPending        FulfillmentStatus = "pending"
	FulfillmentAssigned       FulfillmentStatus = "assigned"
	FulfillmentPicked         FulfillmentStatus = "picked"
	FulfillmentPacked         FulfillmentStatus = "packed"
	FulfillmentShipped        FulfillmentStatus = "shipped"
	FulfillmentOutForDelivery FulfillmentStatus = "out_for_delivery"
	FulfillmentDelivered      FulfillmentStatus = "delivered"
	FulfillmentCancelled      FulfillmentStatus = "cancelled"
	FulfillmentFailed         FulfillmentStatus = "failed"
)

// Order represents an order in the domain layer
type Order struct {
	ID               string
	TenantID         string
	OrgID            string
	CustomerID       *string
	Status           OrderStatus
	PaymentStatus    *string
	PaymentProvider  *string
	PaymentReference *string
	Currency         string
	Subtotal         float64
	DiscountTotal    float64
	ShippingTotal    float64
	TaxTotal         float64
	Total            float64
	IdempotencyKey   *string
	ConfirmedAt      *time.Time
	CancelledAt      *time.Time
	FulfilledAt      *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time

	// Associations
	LineItems []OrderLineItem
}

type OrderLineItem struct {
	ID                string
	OrderID           string
	VariantID         *string
	SKU               *string
	ProductTitle      string
	VariantTitle      *string
	UnitPrice         float64
	Quantity          int
	QuantityFulfilled int
	QuantityReturned  int
	QuantityCancelled int
}

// Order FSM logic
var validTransitions = map[OrderStatus][]OrderStatus{
	StatusDraft:              {StatusPendingPayment, StatusCancelled},
	StatusPendingPayment:     {StatusPaymentAuthorized, StatusFailed, StatusCancelled},
	StatusPaymentAuthorized:  {StatusConfirmed, StatusFailed, StatusCancelled},
	StatusConfirmed:          {StatusProcessing, StatusCancelled},
	StatusProcessing:         {StatusPartiallyFulfilled, StatusFulfilled},
	StatusPartiallyFulfilled: {StatusFulfilled},
	StatusFulfilled:          {StatusDelivered, StatusReturnRequested},
	StatusDelivered:          {StatusCompleted, StatusReturnRequested},
	StatusReturnRequested:    {StatusPartiallyReturned, StatusReturned},
	StatusReturned:           {StatusRefunded},
	StatusCancelled:          {}, // terminal
	StatusRefunded:           {}, // terminal
	StatusCompleted:          {}, // terminal
	StatusFailed:             {}, // terminal
}

func (o *Order) CanTransitionTo(next OrderStatus) bool {
	allowed, ok := validTransitions[o.Status]
	if !ok {
		return false
	}
	for _, status := range allowed {
		if status == next {
			return true
		}
	}
	return false
}

func (o *Order) TransitionTo(next OrderStatus) error {
	if !o.CanTransitionTo(next) {
		return fmt.Errorf("invalid transition from %s to %s", o.Status, next)
	}
	o.Status = next
	return nil
}

// ComputeOrderStatus derives status based on quantities
func (o *Order) ComputeOrderStatus() OrderStatus {
	// Simple version based on plan logic.
	// We'd typically check total quantity vs fulfilled vs returned.
	totalQty := 0
	fulfilledQty := 0
	returnedQty := 0

	for _, item := range o.LineItems {
		totalQty += item.Quantity
		fulfilledQty += item.QuantityFulfilled
		returnedQty += item.QuantityReturned
	}

	if returnedQty == totalQty {
		return StatusReturned
	}
	if returnedQty > 0 {
		return StatusPartiallyReturned
	}

	if fulfilledQty == totalQty {
		return StatusFulfilled
	}
	if fulfilledQty > 0 {
		return StatusPartiallyFulfilled
	}

	return o.Status // Unchanged
}
