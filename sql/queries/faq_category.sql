-- name: GetFaqCategories :many
SELECT * FROM faq_category ORDER BY sort_order, id;

-- name: GetFaqCategoryById :one
SELECT * FROM faq_category WHERE id = ?;

-- name: CreateFaqCategory :execresult
INSERT INTO faq_category (title, sort_order) VALUES (?, ?);

-- name: UpdateFaqCategory :exec
UPDATE faq_category
SET title = ?, sort_order = ?
WHERE id = ?;

-- name: DeleteFaqCategory :exec
DELETE FROM faq_category WHERE id = ?;
