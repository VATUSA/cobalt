-- name: GetEventById :one
SELECT *
FROM event
WHERE id = ?;

-- name: GetEventByIdApproved :one
SELECT *
FROM event
WHERE id = ?
  AND review_status = 'approved';

-- name: GetUpcomingEventsApproved :many
-- Events remain "upcoming" until they finish (end_time), so in-progress events
-- stay visible instead of dropping off the moment they start.
SELECT *
FROM event
WHERE end_time > ?
  AND review_status = 'approved'
ORDER BY start_time ASC
LIMIT ?, ?;

-- name: GetUpcomingEventsAll :many
SELECT *
FROM event
WHERE start_time > ?
ORDER BY start_time ASC
LIMIT ?, ?;

-- name: GetUpcomingEventsForFacility :many
SELECT *
FROM event
WHERE start_time > ?
  AND facility = ?
ORDER BY start_time ASC
LIMIT ?, ?;

-- name: GetUpcomingEventsForFacilityOrPending :many
SELECT *
FROM event
WHERE start_time > ?
  AND (facility = ? OR review_status = 'pending')
ORDER BY start_time ASC
LIMIT ?, ?;

-- name: SetEventReviewStatus :exec
UPDATE event
SET review_status = ?,
    reviewed_by   = ?,
    reviewed_on   = ?
WHERE id = ?;

-- name: CreateEvent :execresult
INSERT INTO event
(title, body, banner_image_url, facility, start_time, end_time, created_at, created_by, updated_at, updated_by, review_status)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'pending');

-- name: UpdateEvent :exec
UPDATE event
SET title            = ?,
    body             = ?,
    banner_image_url = ?,
    facility         = ?,
    start_time       = ?,
    end_time         = ?,
    updated_at       = ?,
    updated_by       = ?
WHERE id = ?;

-- name: DeleteEvent :exec
DELETE
FROM event
WHERE id = ?;
