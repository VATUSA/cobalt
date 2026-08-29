-- name: GetPolicyCategories :many
SELECT * FROM policy_category ORDER BY sort_order, id;

-- name: GetPolicyCategoryById :one
SELECT * FROM policy_category WHERE id = ?;

-- name: CountPolicyDocumentsInCategory :one
SELECT COUNT(*) FROM policy_document WHERE policy_category_id = ?;

-- name: CreatePolicyCategory :execresult
INSERT INTO policy_category (title, sort_order) VALUES (?, ?);

-- name: UpdatePolicyCategory :execresult
UPDATE policy_category
SET title = ?, sort_order = ?
WHERE id = ?;

-- name: DeletePolicyCategory :execresult
DELETE FROM policy_category WHERE id = ?;
