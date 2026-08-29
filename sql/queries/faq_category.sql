-- name: GetFaqCategories :many
SELECT * FROM faq_category ORDER BY sort_order, id;

-- name: GetFaqCategoryById :one
SELECT * FROM faq_category WHERE id = ?;

-- name: CountFaqItemsInCategory :one
SELECT COUNT(*) FROM faq_item WHERE faq_category_id = ?;

-- name: CreateFaqCategory :execresult
INSERT INTO faq_category (title, sort_order) VALUES (?, ?);

-- name: UpdateFaqCategory :execresult
UPDATE faq_category
SET title = ?, sort_order = ?
WHERE id = ?;

-- name: DeleteFaqCategory :execresult
DELETE FROM faq_category WHERE id = ?;
