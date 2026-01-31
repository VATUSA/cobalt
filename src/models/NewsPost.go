package models

import (
	"time"
	"vatusa-cobalt/db"
)

type NewsPostRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type NewsPost struct {
	Title         string `json:"title"`
	Body          string `json:"body"`
	AuthorCID     int    `json:"author_cid"`
	PostTimestamp int64  `json:"post_timestamp"`
	PostDate      string `json:"post_date"`
}

func NewsPostFromDatabase(ent db.NewsPost) NewsPost {
	postTime := time.Unix(ent.PostTime, 0)
	dateString := postTime.Format("2006-01-02")
	return NewsPost{
		Title:         ent.Title,
		Body:          ent.Body,
		AuthorCID:     int(ent.AuthorCid),
		PostTimestamp: ent.PostTime,
		PostDate:      dateString,
	}
}

func NewsPostsFromDatabase(ents []db.NewsPost) []NewsPost {
	posts := make([]NewsPost, len(ents))
	for i, ent := range ents {
		posts[i] = NewsPostFromDatabase(ent)
	}
	return posts
}
