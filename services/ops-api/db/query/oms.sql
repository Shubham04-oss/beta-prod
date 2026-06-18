-- name: GetOrder :one
SELECT * FROM orders
WHERE id = $1 AND tenant_id = $2 AND org_id = $3;

-- name: ListOrders :many
SELECT * FROM orders
WHERE tenant_id = $1 AND org_id = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: UpdateOrderStatus :exec
UPDATE orders
SET status = $1, updated_at = now()
WHERE id = $2 AND tenant_id = $3 AND org_id = $4;

-- name: UpdateOrderPaymentStatus :exec
UPDATE orders
SET payment_status = $1, payment_provider = $2, payment_reference = $3, updated_at = now()
WHERE id = $4 AND tenant_id = $5 AND org_id = $6;

-- name: ReserveInventoryForOrder :one
UPDATE inventory_levels
SET reserved_quantity = reserved_quantity + $1, updated_at = now()
WHERE variant_id = $2
  AND tenant_id = $3
  AND location_id = $4
  AND org_id = $5
  AND (available_quantity - reserved_quantity) >= $1
RETURNING *;

-- name: ReleaseInventoryForOrder :exec
UPDATE inventory_levels
SET reserved_quantity = reserved_quantity - $1, updated_at = now()
WHERE variant_id = $2
  AND tenant_id = $3
  AND location_id = $4
  AND org_id = $5;

-- name: GetCommerceOrderMapping :one
SELECT * FROM commerce_order_mappings
WHERE order_id = $1 AND tenant_id = $2 AND org_id = $3;

-- name: GetCommerceOrderMappingByExternal :one
SELECT * FROM commerce_order_mappings
WHERE connection_id = $1 AND external_order_id = $2 AND tenant_id = $3 AND org_id = $4;
