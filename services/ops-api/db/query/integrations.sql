-- name: CreateTenantIntegration :one
INSERT INTO tenant_integrations (
    tenant_id, org_id, connection_id, category
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetTenantIntegrationByCategory :one
SELECT connection_id FROM tenant_integrations WHERE tenant_id = $1 AND category = $2 LIMIT 1;
