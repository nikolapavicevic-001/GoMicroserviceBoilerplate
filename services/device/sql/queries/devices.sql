-- name: GetDevice :one
SELECT id, name, type, status, created_at, updated_at
FROM devices
WHERE id = $1;

-- name: CreateDevice :one
INSERT INTO devices (name, type, status)
VALUES ($1, $2, $3)
RETURNING id, name, type, status, created_at, updated_at;

-- name: UpdateDevice :one
UPDATE devices
SET name = $2,
    type = $3,
    status = $4,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, name, type, status, created_at, updated_at;

-- name: DeleteDevice :exec
DELETE FROM devices
WHERE id = $1;

-- name: ListDevices :many
SELECT id, name, type, status, created_at, updated_at
FROM devices
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountDevices :one
SELECT COUNT(*) FROM devices;

