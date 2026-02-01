-- name: GetRecentNewsPosts :many
SELECT * FROM news_post order by id desc limit ?;

-- name: CreatePost :exec
INSERT INTO news_post (title, body, author_cid, post_time) VALUES (?, ?, ?, ?);

-- name: GetPostById :one
SELECT * FROM news_post WHERE id = ?;

-- name: DeleteNewsPostById :exec
DELETE FROM news_post WHERE id = ?;