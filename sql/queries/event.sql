-- name: GetEventById :one
SELECT *
FROM event
WHERE id = ?;

-- name: GetUpcomingEvents :many
SELECT *
FROM event
WHERE start_time > ?
ORDER BY start_time DESC
LIMIT ?, ?;

-- name: CreateEvent :exec
INSERT INTO event
(title, body, banner_image_url, facility, start_time, end_time, created_at, created_by, updated_at, updated_by)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateEvent :exec
UPDATE event
SET title            = ?,
    body             = ?,
    banner_image_url = ?,
    facility         = ?,
    start_time       = ?,
    end_time         = ?,
    updated_at       = ?,
    updated_by       =?
WHERE id = ?;

-- name: DeleteEvent :exec
DELETE
FROM event
WHERE id = ?;