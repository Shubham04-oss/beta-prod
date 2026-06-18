# Tier 4 Scenario Test Implementation Plan

## 1. Observation
- `TEST_INFRA.md` specifies 5 "Real-World Application Scenarios (Tier 4)" to be implemented in `/tests/e2e`.
- The current E2E test harness (`tests/e2e/harness/harness.go`) has helper methods `Get`, `Post`, `Put`, `Delete` for API calls, but lacks built-in support for passing custom HTTP headers (such as `X-Tenant-ID` needed for multi-tenancy tests).
- Existing Tier 1 tests (`pim_test.go`, `oms_test.go`, `inventory_test.go`, `unified_sync_test.go`) demonstrate the endpoints and payload structures expected by the mock server (e.g., `PIMProduct`, `Order`, `InventoryStock`, `SyncPayload`).

## 2. Logic Chain
1. To write the 5 scenarios, the Worker must compose sequential API requests that mimic real user workflows, chaining output IDs (like product IDs) to subsequent inputs (like order creation and inventory adjustments).
2. Because the Multi-Tenant scenario requires data isolation testing, the test must be able to send HTTP requests simulating different tenants. This can be achieved by extending `harness.go` with a `DoRequestWithHeaders` function, or by constructing `http.NewRequest` manually within the test.
3. The other 4 scenarios can be built using the existing `harness` methods. They combine features F1-F4 as requested.
4. All tests should be created in a new file `tests/e2e/tier4_test.go` and follow standard Go testing conventions.

## 3. Caveats
- The backend API endpoints might not be fully implemented to enforce multi-tenancy (or they may not use `X-Tenant-ID`). If the backend behaves differently, the test might fail during execution, but the test definition itself remains valid according to typical E2E standards.
- OMS reservation behavior is not fully transparent. The plan explicitly calls the `/inventory/reserve` endpoint to mimic the deduction, as it's safer for black-box testing.

## 4. Conclusion
The implementation plan is as follows:

**1. Update Harness (Optional but recommended)**
Add a `DoRequestWithHeaders` method to `harness/harness.go` to support injecting tenant headers, or use standard `http.Client` directly inside Scenario 5.

**2. Create `tier4_test.go` and implement the 5 scenarios:**

- **Scenario 1: Full Product Lifecycle**
  1. POST `/pim/products` to create.
  2. POST `/inventory/adjust` to set stock.
  3. POST `/unified/sync/push/product` to sync out.
  4. PUT `/pim/products/{id}` to update.
  5. POST `/unified/sync/push/product` to sync changes.
  6. DELETE `/pim/products/{id}`.
  7. GET `/pim/products/{id}` -> Assert 404.

- **Scenario 2: End-to-End Order Fulfillment**
  1. POST `/pim/products` and POST `/inventory/adjust` to initialize.
  2. POST `/oms/orders` to create an order with the product.
  3. POST `/inventory/reserve` to reserve the order quantity.
  4. PUT `/oms/orders/{id}/status` to "shipped".
  5. GET `/inventory/stock/{product_id}` to verify stock deduction.

- **Scenario 3: Bulk Inventory Sync from External**
  1. POST `/pim/products` to create 2 test products.
  2. POST `/unified/webhook` twice simulating `inventory.updated` events from an external system.
  3. GET `/inventory/stock/{id}` for both to verify the local ledger synced properly.

- **Scenario 4: Out-of-Stock Order Handling**
  1. POST `/pim/products` and POST `/inventory/adjust` to a low stock (e.g., 5).
  2. POST `/oms/orders` requesting quantity > 5.
  3. POST `/inventory/reserve` -> Assert failure status (e.g., 400 or 409).
  4. PUT `/oms/orders/{id}/status` -> "cancelled".
  5. GET `/inventory/stock/{id}` -> Assert stock is unchanged.

- **Scenario 5: Multi-Tenant Data Isolation Check**
  1. Use Tenant A headers to create a product.
  2. Use Tenant B headers to attempt GET on Tenant A's product -> Assert 404 Not Found.
  3. Use Tenant B headers to attempt POST `/oms/orders` with Tenant A's product -> Assert 400/404/422.
  4. Create Tenant B's own product and ensure Tenant A cannot access it.

## 5. Verification Method
- The Worker should verify the plan by implementing `tests/e2e/tier4_test.go`.
- Run `go test -v ./tests/e2e -run TestTier4` to ensure the scenarios compile properly and run the requests.
