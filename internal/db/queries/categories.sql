-- name: GetCategory :one
SELECT * FROM categories WHERE id = $1 LIMIT 1;

-- name: GetCategoryByName :one
SELECT * FROM categories WHERE name = $1 LIMIT 1;

-- name: ListCategories :many
SELECT * FROM categories ORDER BY name ASC;

-- name: ListCategoriesPaginated :many
SELECT * FROM categories ORDER BY name ASC LIMIT $1 OFFSET $2;

-- name: CountCategories :one
SELECT COUNT(*) FROM categories;

-- name: CreateCategory :one
INSERT INTO categories (name)
VALUES ($1)
RETURNING *;

-- name: DeleteCategory :exec
DELETE FROM categories WHERE id = $1;
