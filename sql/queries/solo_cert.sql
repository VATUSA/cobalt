-- name: GetActiveSoloCerts :many
SELECT * FROM solo_cert WHERE expires >= CURDATE() ORDER BY expires ASC;

-- name: GetSoloCertById :one
SELECT * FROM solo_cert WHERE id = ?;

-- name: CreateSoloCert :execresult
INSERT INTO solo_cert (cid, facility, position, expires, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?);

-- name: UpdateSoloCert :exec
UPDATE solo_cert
SET facility = ?, position = ?, expires = ?, updated_at = ?
WHERE id = ?;

-- name: DeleteSoloCert :exec
DELETE FROM solo_cert WHERE id = ?;
