-- name: CreateFolder :one
INSERT INTO folders (name, path)
VALUES ($1, $2)
RETURNING *;

-- name: CreateFolderWithParent :one
INSERT INTO folders (name, path, parent_folder_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetFolder :one
SELECT * FROM folders
WHERE id = $1;

-- name: GetFolderByPath :one
SELECT * FROM folders
WHERE path = $1;

-- name: ListFolders :many
SELECT * FROM folders
ORDER BY name;

-- name: ListFoldersPaginated :many
SELECT * FROM folders
ORDER BY name
LIMIT $1 OFFSET $2;

-- name: CountFolders :one
SELECT COUNT(*) FROM folders;

-- name: ListRootFolders :many
SELECT * FROM folders
WHERE parent_folder_id IS NULL
ORDER BY name;

-- name: ListRootFoldersPaginated :many
SELECT * FROM folders
WHERE parent_folder_id IS NULL
ORDER BY name
LIMIT $1 OFFSET $2;

-- name: CountRootFolders :one
SELECT COUNT(*) FROM folders
WHERE parent_folder_id IS NULL;

-- name: SearchFoldersPaginated :many
SELECT * FROM folders
WHERE name ILIKE '%' || @search::text || '%'
ORDER BY name
LIMIT $1 OFFSET $2;

-- name: CountSearchFolders :one
SELECT COUNT(*) FROM folders
WHERE name ILIKE '%' || @search::text || '%';

-- name: SearchRootFoldersPaginated :many
SELECT * FROM folders
WHERE parent_folder_id IS NULL
  AND name ILIKE '%' || @search::text || '%'
ORDER BY name
LIMIT $1 OFFSET $2;

-- name: CountSearchRootFolders :one
SELECT COUNT(*) FROM folders
WHERE parent_folder_id IS NULL
  AND name ILIKE '%' || @search::text || '%';

-- name: ListSubfolders :many
SELECT * FROM folders
WHERE parent_folder_id = $1
ORDER BY name;

-- name: ListSubfoldersPaginated :many
SELECT * FROM folders
WHERE parent_folder_id = $1
ORDER BY name
LIMIT $2 OFFSET $3;

-- name: CountSubfolders :one
SELECT COUNT(*) FROM folders
WHERE parent_folder_id = $1;

-- name: UpdateFolder :one
UPDATE folders
SET name = $2, path = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateFolderParent :one
UPDATE folders
SET parent_folder_id = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteFolder :exec
DELETE FROM folders
WHERE id = $1;

-- name: GetFolderFiles :many
SELECT f.* FROM files f
WHERE f.folder_id = $1
ORDER BY f.file_name;

-- name: GetFolderFilesPaginated :many
SELECT f.* FROM files f
WHERE f.folder_id = $1
ORDER BY f.file_name
LIMIT $2 OFFSET $3;

-- name: CountFolderFiles :one
SELECT COUNT(*) FROM files
WHERE folder_id = $1;

-- name: GetFolderCategories :many
SELECT c.* FROM categories c
INNER JOIN folders_categories fc ON c.id = fc.category_id
WHERE fc.folder_id = $1
ORDER BY c.name;

-- name: GetFolderCategoriesBatch :many
SELECT fc.folder_id, c.*
FROM folders_categories fc
INNER JOIN categories c ON c.id = fc.category_id
WHERE fc.folder_id = ANY(@folder_ids::uuid[])
ORDER BY fc.folder_id, c.name;

-- name: AddFolderCategory :exec
INSERT INTO folders_categories (folder_id, category_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemoveFolderCategory :exec
DELETE FROM folders_categories
WHERE folder_id = $1 AND category_id = $2;

-- name: SetFolderCategories :exec
DELETE FROM folders_categories WHERE folder_id = $1;

-- name: BulkRemoveFolderCategories :exec
DELETE FROM folders_categories WHERE folder_id = ANY(@folder_ids::uuid[]);

-- name: BulkAddFolderCategories :exec
INSERT INTO folders_categories (folder_id, category_id)
SELECT UNNEST(@folder_ids::uuid[]), UNNEST(@category_ids::uuid[])
ON CONFLICT DO NOTHING;

-- name: UpdateFileFolderID :exec
UPDATE files
SET folder_id = $2
WHERE id = $1;
