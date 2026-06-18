-- name: CreateSupplier :one
INSERT INTO suppliers (
    org_id, tenant_id, name, contact_email, payment_terms, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetSupplier :one
SELECT * FROM suppliers 
WHERE id = $1 AND tenant_id = $2;

-- name: GetFirstSupplier :one
SELECT * FROM suppliers 
WHERE tenant_id = $1 LIMIT 1;

-- name: CreatePurchaseOrder :one
INSERT INTO purchase_orders (
    org_id, tenant_id, supplier_id, status, total_amount, currency, expected_delivery_date, notes, created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: CreatePurchaseOrderLineItem :one
INSERT INTO purchase_order_line_items (
    org_id, tenant_id, po_id, variant_id, quantity, unit_price, subtotal
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;
