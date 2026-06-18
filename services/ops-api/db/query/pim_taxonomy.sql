-- name: CreateBrand :one
INSERT INTO brands (org_id, tenant_id, name, description, logo_url)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetBrands :many
SELECT * FROM brands 
WHERE tenant_id = $1 AND deleted_at IS NULL
ORDER BY name ASC;

-- name: CreateCategory :one
INSERT INTO categories (org_id, tenant_id, parent_id, name, slug, description)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetCategories :many
SELECT * FROM categories 
WHERE tenant_id = $1 AND deleted_at IS NULL
ORDER BY name ASC;

