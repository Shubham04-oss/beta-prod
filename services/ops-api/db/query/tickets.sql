-- name: CreateWSTicket :one
INSERT INTO ws_tickets (
    tenant_id,
    org_id,
    user_id,
    role,
    expires_at
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING id;

-- name: ConsumeWSTicket :one
DELETE FROM ws_tickets
WHERE id = $1 AND expires_at > NOW()
RETURNING tenant_id, org_id, user_id, role;

-- name: CleanupWSTickets :exec
DELETE FROM ws_tickets
WHERE expires_at <= NOW();
