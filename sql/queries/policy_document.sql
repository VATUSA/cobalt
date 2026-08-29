-- name: GetAllPolicyDocuments :many
SELECT * FROM policy_document ORDER BY policy_category_id, sort_order, id;

-- name: GetVisiblePolicyDocuments :many
SELECT * FROM policy_document WHERE hidden = false ORDER BY policy_category_id, sort_order, id;

-- name: GetPolicyDocumentById :one
SELECT * FROM policy_document WHERE id = ?;

-- name: CreatePolicyDocument :execresult
INSERT INTO policy_document (policy_category_id, ident, title, summary, document_url, effective_date, hidden, sort_order, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdatePolicyDocument :exec
UPDATE policy_document
SET policy_category_id = ?, ident = ?, title = ?, summary = ?, document_url = ?, effective_date = ?, hidden = ?, sort_order = ?, updated_at = ?
WHERE id = ?;

-- name: DeletePolicyDocument :exec
DELETE FROM policy_document WHERE id = ?;
