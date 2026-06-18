package e2e

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"commerce_modules/tests/e2e/harness"
)

func TestChallenge_ReleaseInventory_SwallowsErrorAndLeaksStock(t *testing.T) {
	h := harness.Setup(t)

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set")
	}
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to db: %v", err)
	}
	defer pool.Close()

	// 1. Create product & two variants
	prodID := uuid.New().String()
	var1ID := uuid.New().String()
	var2ID := uuid.New().String()
	tenantID := "00000000-0000-0000-0000-000000000000"
	orgID := "00000000-0000-0000-0000-000000000000"

	h.Post(t, "/pim/products", map[string]interface{}{
		"id":          prodID,
		"title":       "Test Prod",
		"description": "Test Desc",
		"status":      "active",
	}, nil)

	h.Post(t, "/pim/products/"+prodID+"/variants", map[string]interface{}{
		"id":       var1ID,
		"sku":      "V1-" + uuid.New().String(),
		"price":    10.0,
		"currency": "USD",
	}, nil)

	h.Post(t, "/pim/products/"+prodID+"/variants", map[string]interface{}{
		"id":       var2ID,
		"sku":      "V2-" + uuid.New().String(),
		"price":    20.0,
		"currency": "USD",
	}, nil)

	// 2. Add inventory for both
	h.Post(t, "/inventory/adjust", map[string]interface{}{
		"product_id": var1ID,
		"quantity":   5,
	}, nil)
	h.Post(t, "/inventory/adjust", map[string]interface{}{
		"product_id": var2ID,
		"quantity":   5,
	}, nil)

	// 3. Create an order with both items
	var orderResp map[string]interface{}
	statusCode := h.Post(t, "/oms/orders", map[string]interface{}{
		"customer_id": uuid.New().String(),
		"currency":    "USD",
		"items": []map[string]interface{}{
			{"product_id": var1ID, "quantity": 2},
			{"product_id": var2ID, "quantity": 2},
		},
	}, &orderResp)
	
	if statusCode != http.StatusCreated {
		t.Fatalf("Failed to create order, status %d", statusCode)
	}
	orderID := orderResp["id"].(string)

	// At this point, var1 has 2 reserved, var2 has 2 reserved.
	// Let's manually screw up var2's reserved quantity in DB to cause ReleaseStock to fail.
	var2UUID := uuid.MustParse(var2ID)
	// We need the md5 variant id
	var2Hash := uuid.NewMD5(uuid.NameSpaceOID, []byte(var2UUID.String()))
	
	_, err = pool.Exec(context.Background(), `
		UPDATE inventory_levels SET reserved_quantity = 0 WHERE variant_id = $1
	`, var2Hash)
	if err != nil {
		t.Fatalf("Failed to mutate DB: %v", err)
	}

	// 4. Cancel the order. This will call ReleaseInventory.
	// It should release var1 (success), then fail on var2 (because 0 < 2).
	// Because of the bug, the error is returned to OMS, so OMS rolls back the order status to Pending!
	// BUT var1's inventory was already released and is NOT rolled back!
	statusCode = h.Post(t, "/oms/orders/"+orderID+"/cancel", nil, nil)
	if statusCode == http.StatusOK {
		t.Fatalf("Expected cancellation to fail, but it succeeded!")
	}

	// 5. Verify the order is still Pending
	var checkOrderResp map[string]interface{}
	h.Get(t, "/oms/orders/"+orderID, &checkOrderResp)
	if checkOrderResp["status"] != "pending" {
		t.Fatalf("Expected order status to be 'pending' due to rollback, got %v", checkOrderResp["status"])
	}

	// 6. Verify var1's reserved quantity. It should be 2, but it will be 0 due to the bug!
	var avail, resrv int
	var1UUID := uuid.MustParse(var1ID)
	var1Hash := uuid.NewMD5(uuid.NameSpaceOID, []byte(var1UUID.String()))
	err = pool.QueryRow(context.Background(), `
		SELECT available_quantity, reserved_quantity FROM inventory_levels WHERE variant_id = $1
	`, var1Hash).Scan(&avail, &resrv)
	
	if err != nil {
		t.Fatalf("Failed to get var1 stock: %v", err)
	}

	if resrv != 2 {
		t.Fatalf("CHALLENGE SUCCESS: Bug found! Order is still pending, but var1's reserved stock was released and not rolled back. Expected reserved=2, got reserved=%d", resrv)
	}
}
