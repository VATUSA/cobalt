package models

import (
	"time"
	"vatusa-cobalt/config"
	"vatusa-cobalt/db"
)

type EventRequest struct {
	Title          string `json:"title"`
	Body           string `json:"body"`
	BannerImageURL string `json:"banner_image_url"`
	Facility       string `json:"facility"`
	StartTimestamp string `json:"start_timestamp"`
	EndTimestamp   string `json:"end_timestamp"`
}

type EventReviewRequest struct {
	Status string `json:"status"`
}

type Event struct {
	Id             int     `json:"id"`
	Title          string  `json:"title"`
	Body           string  `json:"body"`
	BannerImageURL string  `json:"banner_image_url"`
	Facility       string  `json:"facility"`
	StartTimestamp string  `json:"start_timestamp"`
	EndTimestamp   string  `json:"end_timestamp"`
	ReviewStatus   *string `json:"review_status"`
	ReviewedBy     *int32  `json:"reviewed_by"`
	ReviewedOn     *int64  `json:"reviewed_on"`
}

func EventFromDatabase(ent db.Event) Event {
	startTime := time.Unix(ent.StartTime, 0)
	endTime := time.Unix(ent.EndTime, 0)
	var reviewStatus *string
	if ent.ReviewStatus.Valid {
		reviewStatus = &ent.ReviewStatus.String
	}
	var reviewedBy *int32
	if ent.ReviewedBy.Valid {
		reviewedBy = &ent.ReviewedBy.Int32
	}
	var reviewedOn *int64
	if ent.ReviewedOn.Valid {
		reviewedOn = &ent.ReviewedOn.Int64
	}
	return Event{
		Id:             int(ent.ID),
		Title:          ent.Title,
		Body:           ent.Body,
		BannerImageURL: ent.BannerImageUrl,
		Facility:       ent.Facility,
		StartTimestamp: startTime.Format(config.TimestampFormat),
		EndTimestamp:   endTime.Format(config.TimestampFormat),
		ReviewStatus:   reviewStatus,
		ReviewedBy:     reviewedBy,
		ReviewedOn:     reviewedOn,
	}
}

func EventsFromDatabase(ents []db.Event) []Event {
	events := make([]Event, len(ents))
	for i, ent := range ents {
		events[i] = EventFromDatabase(ent)
	}
	return events
}
