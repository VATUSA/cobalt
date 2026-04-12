package models

import (
	"time"
	"vatusa-cobalt/db"
	"vatusa-cobalt/dbconn"
)

type TransferRequest struct {
	Id           int64     `json:"id"`
	Cid          int64     `json:"cid"`
	FromFacility string    `json:"from_facility"`
	ToFacility   string    `json:"to_facility"`
	Reason       string    `json:"reason"`
	CreatedAt    time.Time `json:"created_at"`
	Status       string    `json:"status"`
}

type TransferRequestCombined struct {
	TransferRequest TransferRequest `json:"transfer_request"`
	User            User            `json:"user"`
}

func TransferRequestsCombinedFromDatabase(requests []dbconn.TransferRequestCombined, canSeeSensitiveFields bool) []TransferRequestCombined {
	var out []TransferRequestCombined
	for _, request := range requests {
		out = append(out, TransferRequestCombinedFromDatabase(request, canSeeSensitiveFields))
	}
	return out
}

func TransferRequestCombinedFromDatabase(request dbconn.TransferRequestCombined, canSeeSensitiveFields bool) TransferRequestCombined {
	return TransferRequestCombined{
		TransferRequest: TransferRequestFromDatabase(request.TransferRequest),
		User:            UserFromDatabase(request.CombinedUserRow, canSeeSensitiveFields),
	}
}

func TransferRequestsFromDatabase(requests []db.TransferRequest) []TransferRequest {
	var out []TransferRequest
	for _, request := range requests {
		out = append(out, TransferRequestFromDatabase(request))
	}
	return out
}

func TransferRequestFromDatabase(request db.TransferRequest) TransferRequest {
	return TransferRequest{
		Id:           request.ID,
		Cid:          request.Cid,
		FromFacility: request.FromFacility,
		ToFacility:   request.ToFacility,
		Reason:       request.Reason,
		CreatedAt:    request.CreatedAt,
		Status:       request.Status,
	}
}
