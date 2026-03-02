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

type Event struct {
	Id             int    `json:"id"`
	Title          string `json:"title"`
	Body           string `json:"body"`
	BannerImageURL string `json:"banner_image_url"`
	Facility       string `json:"facility"`
	StartTimestamp string `json:"start_timestamp"`
	EndTimestamp   string `json:"end_timestamp"`
}

func EventFromDatabase(ent db.Event) Event {
	startTime := time.Unix(ent.StartTime, 0)
	startTimeString := startTime.Format(config.TimestampFormat)
	endTime := time.Unix(ent.EndTime, 0)
	endTimeString := endTime.Format(config.TimestampFormat)
	return Event{
		Id:             int(ent.ID),
		Title:          ent.Title,
		Body:           ent.Body,
		BannerImageURL: ent.BannerImageUrl,
		Facility:       ent.Facility,
		StartTimestamp: startTimeString,
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
