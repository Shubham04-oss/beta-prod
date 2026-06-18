-- name: ListTenantMembers :many
SELECT id, email, role, created_at
FROM users
WHERE tenant_id = $1
ORDER BY created_at DESC;

-- name: CreateTenantMember :one
INSERT INTO users (
  org_id,
  tenant_id,
  email,
  role
) VALUES (
  $1, $2, $3, $4
)
RETURNING id, email, role, created_at;
