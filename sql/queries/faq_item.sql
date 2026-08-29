-- name: GetAllFaqItems :many
SELECT * FROM faq_item ORDER BY faq_category_id, sort_order, id;

-- name: GetFaqItemById :one
SELECT * FROM faq_item WHERE id = ?;

-- name: CreateFaqItem :execresult
INSERT INTO faq_item (faq_category_id, question, answer, sort_order) VALUES (?, ?, ?, ?);

-- name: UpdateFaqItem :exec
UPDATE faq_item
SET faq_category_id = ?, question = ?, answer = ?, sort_order = ?
WHERE id = ?;

-- name: DeleteFaqItem :exec
DELETE FROM faq_item WHERE id = ?;
