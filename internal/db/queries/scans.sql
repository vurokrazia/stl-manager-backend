-- name: GetScan :one
SELECT * FROM scans WHERE id = $1 LIMIT 1;

-- name: ListScans :many
SELECT * FROM scans
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountScans :one
SELECT COUNT(*) FROM scans;

-- name: CreateScan :one
INSERT INTO scans (status, found, processed, progress)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateScan :one
UPDATE scans
SET status = $2, found = $3, processed = $4, progress = $5, error = $6, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteScan :exec
DELETE FROM scans WHERE id = $1;
