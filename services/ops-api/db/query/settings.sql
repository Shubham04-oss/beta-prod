-- name: GetTenantSettings :one
SELECT * FROM tenant_settings 
WHERE tenant_id = $1 AND org_id = $2 LIMIT 1;

-- name: UpsertTenantSettings :one
INSERT INTO tenant_settings (
    org_id, tenant_id, inventory_allocation_model, auto_po_enabled, default_low_stock_threshold, costing_method, updated_by, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP
) ON CONFLICT (tenant_id) DO UPDATE
SET 
    inventory_allocation_model = EXCLUDED.inventory_allocation_model,
    auto_po_enabled = EXCLUDED.auto_po_enabled,
    default_low_stock_threshold = EXCLUDED.default_low_stock_threshold,
    costing_method = EXCLUDED.costing_method,
    updated_by = EXCLUDED.updated_by,
    updated_at = CURRENT_TIMESTAMP
WHERE tenant_settings.org_id = EXCLUDED.org_id
RETURNING *;
