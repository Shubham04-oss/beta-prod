-- name: GetAuditEventsByTenant :many
SELECT id, actor_email, action, entity_type, entity_id, details, ip_address, created_at 
FROM audit_events
WHERE tenant_id = $1 AND org_id = $2
ORDER BY created_at DESC
LIMIT 50;

-- name: InsertAuditEvent :one
INSERT INTO audit_events (
    org_id,
    tenant_id,
    actor_email,
    action,
    entity_type,
    entity_id,
    details,
    ip_address
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING id;
