-- name: GetActiveSoloCerts :many
SELECT * FROM solo_cert WHERE expires >= UTC_DATE() ORDER BY expires ASC;

-- name: GetSoloCertById :one
SELECT * FROM solo_cert WHERE id = ?;

-- name: GetActiveSoloCertForCidPosition :many
SELECT * FROM solo_cert WHERE cid = ? AND position = ? AND expires >= UTC_DATE();

-- name: CreateSoloCert :execresult
INSERT INTO solo_cert (cid, facility, position, expires, created_by_cid, updated_by_cid, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateSoloCert :execresult
UPDATE solo_cert
SET position = ?, expires = ?, updated_by_cid = ?, updated_at = ?
WHERE id = ?;

-- name: DeleteSoloCert :execresult
DELETE FROM solo_cert WHERE id = ?;
