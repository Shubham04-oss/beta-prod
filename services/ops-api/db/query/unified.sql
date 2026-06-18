-- name: GetCommerceConnections :many
SELECT * FROM commerce_connections
WHERE tenant_id = $1 AND org_id = $2 AND status = 'ACTIVE' AND deleted_at IS NULL;

-- name: GetTenantByConnectionID :one
SELECT org_id, tenant_id FROM commerce_connections
WHERE unified_connection_id = $1 AND status = 'ACTIVE' AND deleted_at IS NULL
LIMIT 1;

-- name: CreateCommerceConnection :one
INSERT INTO commerce_connections (
  org_id, tenant_id, unified_connection_id, provider
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;
