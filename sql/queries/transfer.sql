-- name: CreateTransferRequest :execresult
INSERT INTO transfer_request (cid, from_facility, to_facility, reason, created_at, status)
VALUES (?, ?, ?, ?, NOW(), 'pending');

-- name: UpdateTransferRequestStatus :exec
UPDATE transfer_request SET status = ? WHERE id = ?;

-- name: DeleteTransferRequest :exec
DELETE FROM transfer_request WHERE id = ?;

-- name: CreateTransferHistoryRecord :execresult
INSERT INTO transfer_history (cid, from_facility, to_facility, reason, requested_at, completed_at, approver_cid)
VALUES (?, ?, ?, ?, ?, NOW(), ?);

-- name: GetPendingTransferRequestsForFacility :many
SELECT id, cid, from_facility, to_facility, reason, created_at, status
FROM transfer_request
WHERE to_facility = ?;

-- name: GetTransferRequestsForCID :many
SELECT id, cid, from_facility, to_facility, reason, created_at, status
FROM transfer_request
WHERE cid = ?;

-- name: GetTransferHistoryForCID :many
SELECT id, cid, from_facility, to_facility, reason, requested_at, completed_at, approver_cid
FROM transfer_history
WHERE cid = ?;

-- name: GetTransferRequestById :one
SELECT *
FROM transfer_request
WHERE id = ?;