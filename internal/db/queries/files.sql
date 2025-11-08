-- name: GetFile :one
SELECT * FROM files WHERE id = $1 LIMIT 1;

-- name: GetFileByPath :one
SELECT * FROM files WHERE path = $1 LIMIT 1;

-- name: ListFiles :many
SELECT * FROM files
ORDER BY file_name ASC
LIMIT $1 OFFSET $2;

-- name: SearchFiles :many
SELECT
  f.*,
  similarity(f.file_name, $1) as sim
FROM files f
WHERE
  ($1 = '' OR f.file_name % $1 OR f.path % $1)
  AND ($2 = '' OR f.type = $2)
ORDER BY sim DESC, f.file_name ASC
LIMIT $3 OFFSET $4;

-- name: CreateFile :one
INSERT INTO files (path, file_name, type, size, modified_at, sha256)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateFile :one
UPDATE files
SET file_name = $2, type = $3, size = $4, modified_at = $5, sha256 = $6, updated_at = now()
WHERE path = $1
RETURNING *;

-- name: UpsertFile :one
INSERT INTO files (path, file_name, type, size, modified_at, sha256, folder_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (path)
DO UPDATE SET
  file_name = EXCLUDED.file_name,
  type = EXCLUDED.type,
  size = EXCLUDED.size,
  modified_at = EXCLUDED.modified_at,
  sha256 = EXCLUDED.sha256,
  folder_id = EXCLUDED.folder_id,
  updated_at = now()
RETURNING *;

-- name: DeleteFile :exec
DELETE FROM files WHERE id = $1;

-- name: CountFiles :one
SELECT COUNT(*) FROM files;

-- name: CountFilesByType :one
SELECT COUNT(*) FROM files WHERE type = $1;

-- name: ListRootFiles :many
SELECT * FROM files
WHERE folder_id IS NULL
ORDER BY file_name ASC;

-- name: ListAllFiles :many
SELECT * FROM files
ORDER BY file_name ASC;

-- name: ListAllFilesPaginated :many
SELECT * FROM files
ORDER BY file_name ASC
LIMIT $1 OFFSET $2;

-- name: ListRootFilesPaginated :many
SELECT * FROM files
WHERE folder_id IS NULL
ORDER BY file_name ASC
LIMIT $1 OFFSET $2;

-- name: CountRootFiles :one
SELECT COUNT(*) FROM files WHERE folder_id IS NULL;
