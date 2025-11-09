-- name: GetFileCategories :many
SELECT c.* FROM categories c
INNER JOIN files_categories fc ON fc.category_id = c.id
WHERE fc.file_id = $1
ORDER BY c.name ASC;

-- name: GetCategoriesBatch :many
SELECT fc.file_id, c.*
FROM files_categories fc
INNER JOIN categories c ON c.id = fc.category_id
WHERE fc.file_id = ANY(@file_ids::uuid[])
ORDER BY fc.file_id, c.name;

-- name: AddFileCategory :exec
INSERT INTO files_categories (file_id, category_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemoveFileCategory :exec
DELETE FROM files_categories
WHERE file_id = $1 AND category_id = $2;

-- name: RemoveAllFileCategories :exec
DELETE FROM files_categories WHERE file_id = $1;

-- name: BulkRemoveFileCategories :exec
DELETE FROM files_categories WHERE file_id = ANY(@file_ids::uuid[]);

-- name: BulkAddFileCategories :exec
INSERT INTO files_categories (file_id, category_id)
SELECT UNNEST(@file_ids::uuid[]), UNNEST(@category_ids::uuid[])
ON CONFLICT DO NOTHING;

-- name: GetFilesByCategory :many
SELECT f.* FROM files f
INNER JOIN files_categories fc ON fc.file_id = f.id
INNER JOIN categories c ON c.id = fc.category_id
WHERE c.name = $1
ORDER BY f.file_name ASC
LIMIT $2 OFFSET $3;
