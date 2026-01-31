-- name: GetRecentNewsPosts :many
SELECT * FROM news_post order by id desc limit ?;

-- name: CreatePost :exec
INSERT INTO news_post (title, body, author_cid, post_time) VALUES (?, ?, ?, ?);