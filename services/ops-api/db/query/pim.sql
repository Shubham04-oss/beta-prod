-- name: CreateProductType :one
INSERT INTO product_types (
    id, org_id, tenant_id, name, description
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: CreateAttribute :one
INSERT INTO attributes (
    id, org_id, tenant_id, name, slug, type
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: CreateAttributeValue :one
INSERT INTO attribute_values (
    id, org_id, tenant_id, attribute_id, value, slug
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: LinkProductTypeAttribute :exec
INSERT INTO product_type_attributes (
    product_type_id, attribute_id, org_id, tenant_id
) VALUES (
    $1, $2, $3, $4
);

-- name: CreateProduct :one
INSERT INTO products (
    id, org_id, tenant_id, title, description, status, product_type_id, created_by,
    short_description, long_description, category, brand, tags,
    launch_date, discontinue_date, warranty_period,
    is_taxable, is_returnable, requires_serial_number,
    seo_title, seo_description, seo_keywords, data_quality_score
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8,
    $9, $10, $11, $12, $13,
    $14, $15, $16,
    $17, $18, $19,
    $20, $21, $22, $23
) RETURNING id, org_id, tenant_id, title, description, status, product_type_id, created_by, updated_by, created_at, updated_at, deleted_at,
    short_description, long_description, category, brand, tags,
    launch_date, discontinue_date, warranty_period,
    is_taxable, is_returnable, requires_serial_number,
    seo_title, seo_description, seo_keywords, data_quality_score;

-- name: CreateProductVariant :one
INSERT INTO product_variants (
    id, org_id, tenant_id, product_id, sku, barcode, price, currency, created_by, data_quality_score
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: LinkVariantAttributeValue :exec
INSERT INTO variant_attribute_values (
    variant_id, attribute_value_id, org_id, tenant_id
) VALUES (
    $1, $2, $3, $4
);

-- name: CreateProductMedia :one
INSERT INTO product_media (
    id, org_id, tenant_id, product_id, variant_id, url, alt_text, sort_order
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetProductWithVariants :many
SELECT 
    p.id as product_id,
    p.title as product_title,
    p.data_quality_score as product_score,
    v.id as variant_id,
    v.sku,
    v.price,
    av.value as attribute_value,
    a.name as attribute_name
FROM products p
LEFT JOIN product_variants v ON v.product_id = p.id
LEFT JOIN variant_attribute_values vav ON vav.variant_id = v.id
LEFT JOIN attribute_values av ON av.id = vav.attribute_value_id
LEFT JOIN attributes a ON a.id = av.attribute_id
WHERE p.id = $1 AND p.tenant_id = current_setting('app.current_tenant', true)::uuid
  AND p.deleted_at IS NULL AND v.deleted_at IS NULL;

-- name: ListProductsForUCP :many
SELECT 
    p.id as product_id,
    p.title,
    p.description,
    p.brand,
    p.category,
    pv.id as variant_id,
    pv.sku,
    pv.barcode as gtin,
    pv.price,
    pv.currency,
    pm.url as image_url
FROM products p
LEFT JOIN product_variants pv ON p.id = pv.product_id
LEFT JOIN product_media pm ON p.id = pm.product_id
WHERE p.tenant_id = $1 AND p.status = 'ACTIVE' AND p.deleted_at IS NULL AND pv.deleted_at IS NULL;

-- name: ListProductsByTenant :many
SELECT 
    p.id, 
    p.title, 
    p.category, 
    p.status, 
    p.updated_at
FROM products p
WHERE p.tenant_id = $1 AND p.deleted_at IS NULL
ORDER BY p.updated_at DESC
LIMIT 100;

-- name: GetProduct :one
SELECT *
FROM products
WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL;

-- name: UpdateProduct :one
UPDATE products
SET 
    title = COALESCE(NULLIF($3::text, ''), title),
    description = COALESCE(NULLIF($4::text, ''), description),
    category = COALESCE(NULLIF($5::text, ''), category),
    updated_at = NOW()
WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteProduct :exec
UPDATE products
SET deleted_at = NOW()
WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL;

-- name: GetPIMStatsByTenant :one
SELECT 
    COUNT(DISTINCT p.id) as total_products,
    COALESCE(SUM(CASE WHEN il.available_quantity > 0 AND il.available_quantity < 20 THEN 1 ELSE 0 END), 0)::int as low_stock_variants,
    COALESCE(SUM(CASE WHEN il.available_quantity <= 0 THEN 1 ELSE 0 END), 0)::int as out_of_stock_variants,
    COALESCE(SUM(il.available_quantity * pv.cost_price), 0)::numeric as total_inventory_value
FROM products p
LEFT JOIN product_variants pv ON pv.product_id = p.id AND pv.deleted_at IS NULL
LEFT JOIN inventory_levels il ON il.variant_id = pv.id
WHERE p.tenant_id = $1 AND p.deleted_at IS NULL;

-- name: GetTopLowStockProducts :many
SELECT 
    p.id as product_id,
    p.title as product_title,
    il.available_quantity as stock_left
FROM products p
JOIN product_variants pv ON pv.product_id = p.id
JOIN inventory_levels il ON il.variant_id = pv.id
WHERE p.tenant_id = $1 
  AND p.deleted_at IS NULL
  AND pv.deleted_at IS NULL
  AND il.available_quantity < 20
ORDER BY il.available_quantity ASC
LIMIT 5;

-- name: CreateInventoryLedgerEntry :one
INSERT INTO inventory_ledger (
    org_id, tenant_id, variant_id, location_id, transaction_type, quantity_delta, unit_cost, notes, created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: ReserveInventory :exec
UPDATE inventory_levels
SET 
    available_quantity = available_quantity - $3,
    reserved_quantity = reserved_quantity + $3,
    updated_at = CURRENT_TIMESTAMP
WHERE tenant_id = $1 AND variant_id = $2 AND location_id = $4;

-- name: FulfillReservation :exec
UPDATE inventory_levels
SET 
    reserved_quantity = reserved_quantity - $3,
    updated_at = CURRENT_TIMESTAMP
WHERE tenant_id = $1 AND variant_id = $2 AND location_id = $4;

-- name: GetAttributes :many
SELECT * FROM attributes 
WHERE tenant_id = $1 AND deleted_at IS NULL
ORDER BY name ASC;

-- name: GetProductTypes :many
SELECT * FROM product_types 
WHERE tenant_id = $1 AND deleted_at IS NULL
ORDER BY name ASC;

-- name: GetProductMedia :many
SELECT * FROM product_media 
WHERE tenant_id = $1 AND (product_id = $2 OR variant_id = $3)
ORDER BY sort_order ASC;

-- name: GetAuditEvents :many
SELECT * FROM audit_events 
WHERE tenant_id = $1 
ORDER BY created_at DESC
LIMIT 100;

-- name: GetValidationIssues :many
SELECT * FROM validation_issues 
WHERE tenant_id = $1 AND resolved = false
ORDER BY created_at DESC;

-- name: CreateAttributeGroup :one
INSERT INTO attribute_groups (
    org_id, tenant_id, name, description
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetAttributeGroups :many
SELECT * FROM attribute_groups 
WHERE tenant_id = $1 
ORDER BY created_at DESC;

-- name: CreateBulkJob :one
INSERT INTO bulk_jobs (
    org_id, tenant_id, job_type, status, payload_json, created_by
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetBulkJobs :many
SELECT * FROM bulk_jobs 
WHERE tenant_id = $1 AND job_type = $2
ORDER BY created_at DESC;



