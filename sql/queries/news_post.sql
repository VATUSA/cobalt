-- name: GetRecentNewsPosts :many
SELECT * FROM news_post order by id desc limit ?;

-- name: CreatePost :execresult
INSERT INTO news_post (title, body, author_cid, post_time) VALUES (?, ?, ?, ?);

-- name: UpdatePost :exec
UPDATE news_post
SET title = ?, body = ?, edit_time = ?
WHERE id = ?;

-- name: GetPostById :one
SELECT * FROM news_post WHERE id = ?;

-- name: DeleteNewsPostById :exec
DELETE FROM news_post WHERE id = ?;

-- name: GetNewsPostsPage :many
SELECT *
FROM news_post
ORDER BY id DESC
LIMIT ?, ?;