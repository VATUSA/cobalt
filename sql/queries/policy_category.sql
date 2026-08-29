-- name: GetPolicyCategories :many
SELECT * FROM policy_category ORDER BY sort_order, id;

-- name: GetPolicyCategoryById :one
SELECT * FROM policy_category WHERE id = ?;

-- name: CreatePolicyCategory :execresult
INSERT INTO policy_category (title, sort_order) VALUES (?, ?);

-- name: UpdatePolicyCategory :exec
UPDATE policy_category
SET title = ?, sort_order = ?
WHERE id = ?;

-- name: DeletePolicyCategory :exec
DELETE FROM policy_category WHERE id = ?;
