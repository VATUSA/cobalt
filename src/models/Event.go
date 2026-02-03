package models

import (
	"time"
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

type Event struct {
	Id             int    `json:"id"`
	Title          string `json:"title"`
	Body           string `json:"body"`
	BannerImageURL string `json:"banner_image_url"`
	Facility       string `json:"facility"`
	StartTime      int    `json:"start_time"`
	StartTimestamp string `json:"start_timestamp"`
	EndTime        int    `json:"end_time"`
	EndTimestamp   string `json:"end_timestamp"`
}

func EventFromDatabase(ent db.Event) Event {
	startTime := time.Unix(ent.StartTime, 0)
	startTimeString := startTime.Format("2006-01-02 15:04:05")
	endTime := time.Unix(ent.EndTime, 0)
	endTimeString := endTime.Format("2006-01-02 15:04:05")
	return Event{
		Id:             int(ent.ID),
		Title:          ent.Title,
		Body:           ent.Body,
		BannerImageURL: ent.BannerImageUrl,
		Facility:       ent.Facility,
		StartTime:      int(ent.StartTime),
		StartTimestamp: startTimeString,
		EndTime:        int(ent.EndTime),
		EndTimestamp:   endTimeString,
	}
}

func EventsFromDatabase(ents []db.Event) []Event {
	events := make([]Event, len(ents))
	for i, ent := range ents {
		events[i] = EventFromDatabase(ent)
	}
	return events
}
