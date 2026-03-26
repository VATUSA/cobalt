package models

import (
	"strings"
	"vatusa-cobalt/db"
)

type User struct {
	CID          int           `json:"cid"`
	NetworkUser  *NetworkUser  `json:"network_user"`
	DivisionUser *DivisionUser `json:"division_user"`
}

type NetworkUser struct {
	FirstName      *string `json:"first_name"`
	LastName       *string `json:"last_name"`
	Email          *string `json:"email"`
	Rating         int     `json:"rating"`
	Region         string  `json:"region"`
	Division       string  `json:"division"`
	SubDivision    *string `json:"subdivision"`
	PilotRating    int     `json:"pilot_rating"`
	MilitaryRating int     `json:"military_rating"`
}

type DivisionUser struct {
	DisplayName            string   `json:"display_name"`
	ControllerRating       int      `json:"controller_rating"`
	InstructorRating       int      `json:"instructor_rating"`
	Facility               string   `json:"facility"`
	VisitingFacilities     []string `json:"visiting_facilities"`
	DiscordId              *string  `json:"discord_id"`
	LastPromotionTimestamp *int64   `json:"last_promotion_timestamp"`
	LastTransferTimestamp  *int64   `json:"last_transfer_timestamp"`
}

func UserFromDatabase(user db.GetCombinedUserByCIDRow, canSeeSensitiveFields bool) User {
	output := User{
		CID: int(user.Cid),
		NetworkUser: &NetworkUser{
			FirstName:      nil,
			LastName:       nil,
			Email:          nil,
			Rating:         int(user.Rating),
			Region:         user.RegionID,
			Division:       user.DivisionID,
			SubDivision:    nil,
			PilotRating:    int(user.Pilotrating),
			MilitaryRating: int(user.Militaryrating),
		},
		DivisionUser: nil,
	}
	if user.SubdivisionID.Valid {
		output.NetworkUser.SubDivision = &user.SubdivisionID.String
	}
	if canSeeSensitiveFields {
		output.NetworkUser.FirstName = &user.NameFirst
		output.NetworkUser.LastName = &user.NameLast
		output.NetworkUser.Email = &user.Email
	}
	output.DivisionUser = &DivisionUser{
		DisplayName:            user.DisplayName,
		ControllerRating:       int(user.ControllerRating),
		InstructorRating:       int(user.InstructorRating),
		Facility:               user.Facility,
		VisitingFacilities:     []string{},
		DiscordId:              nil,
		LastPromotionTimestamp: nil,
		LastTransferTimestamp:  nil,
	}
	if user.VisitingFacilities.Valid {
		visitingFacilities := strings.Split(user.VisitingFacilities.String, ",")
		output.DivisionUser.VisitingFacilities = visitingFacilities
	}
	if user.DiscordID != "" {
		output.DivisionUser.DiscordId = &user.DiscordID
	}
	if user.LastPromotionTime.Valid {
		t := user.LastPromotionTime.Time.Unix()
		output.DivisionUser.LastPromotionTimestamp = &t
	}
	if user.LastTransferTime.Valid {
		t := user.LastTransferTime.Time.Unix()
		output.DivisionUser.LastTransferTimestamp = &t
	}
	return output
}
