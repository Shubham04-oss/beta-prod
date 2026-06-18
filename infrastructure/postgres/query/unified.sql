-- name: GetCommerceConnections :many
SELECT * FROM commerce_connections WHERE tenant_id = $1 AND status = 'ACTIVE';

-- name: RecordSyncFailure :one
INSERT INTO sync_failures_dlq (
    org_id, tenant_id, connection_id, entity_type, entity_id, payload, error_message, retry_count, status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;
