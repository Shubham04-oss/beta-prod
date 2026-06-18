-- name: CreateSalesChannel :one
INSERT INTO sales_channels (
    org_id, tenant_id, name, currency, is_active
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: ListSalesChannels :many
SELECT * FROM sales_channels
WHERE tenant_id = $1 AND org_id = $2
ORDER BY created_at DESC;

-- name: CreateOrUpdateProductChannelListing :one
INSERT INTO product_channel_listings (
    product_id, channel_id, org_id, tenant_id, is_published, published_at, override_title, override_description, channel_price
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
ON CONFLICT (product_id, channel_id) DO UPDATE SET
    is_published = EXCLUDED.is_published,
    published_at = EXCLUDED.published_at,
    override_title = EXCLUDED.override_title,
    override_description = EXCLUDED.override_description,
    channel_price = EXCLUDED.channel_price,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: GetProductWithChannelOverrides :one
SELECT 
    p.id as product_id,
    COALESCE(pcl.override_title, p.title)::varchar as title,
    COALESCE(pcl.override_description, p.description)::text as description,
    COALESCE(pcl.channel_price, pv.price)::numeric as price,
    pcl.is_published,
    pcl.published_at
FROM products p
LEFT JOIN product_variants pv ON p.id = pv.product_id -- Simplification: grabbing first variant price
LEFT JOIN product_channel_listings pcl ON p.id = pcl.product_id AND pcl.channel_id = $2
WHERE p.id = $1 AND p.tenant_id = $3 AND p.org_id = $4
LIMIT 1;
