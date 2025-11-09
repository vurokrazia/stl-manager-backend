-- name: GetCategory :one
SELECT * FROM categories
WHERE id = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: GetCategoryByName :one
SELECT * FROM categories
WHERE name = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: ListCategories :many
SELECT * FROM categories
WHERE deleted_at IS NULL
ORDER BY name ASC;

-- name: ListCategoriesPaginated :many
SELECT * FROM categories
WHERE deleted_at IS NULL
ORDER BY name ASC
LIMIT $1 OFFSET $2;

-- name: CountCategories :one
SELECT COUNT(*) FROM categories
WHERE deleted_at IS NULL;

-- name: SearchCategoriesPaginated :many
SELECT * FROM categories
WHERE deleted_at IS NULL
  AND name ILIKE '%' || @search::text || '%'
ORDER BY name ASC
LIMIT $1 OFFSET $2;

-- name: CountSearchCategories :one
SELECT COUNT(*) FROM categories
WHERE deleted_at IS NULL
  AND name ILIKE '%' || @search::text || '%';

-- name: CreateCategory :one
INSERT INTO categories (name)
VALUES ($1)
RETURNING *;

-- name: UpdateCategory :one
UPDATE categories
SET name = $2
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteCategory :exec
UPDATE categories
SET deleted_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: RestoreCategory :exec
UPDATE categories
SET deleted_at = NULL
WHERE id = $1;

-- name: DeleteCategory :exec
DELETE FROM categories WHERE id = $1;
