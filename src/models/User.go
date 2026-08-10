package models

import (
	"strings"
	"vatusa-cobalt/acl"
	"vatusa-cobalt/db"
)

type UserRole struct {
	Facility  string `json:"facility"`
	Role      string `json:"role"`
	GrantedAt int64  `json:"granted_at"`
}

type User struct {
	CID          int           `json:"cid"`
	NetworkUser  *NetworkUser  `json:"network_user"`
	DivisionUser *DivisionUser `json:"division_user"`
	Roles        []UserRole    `json:"roles"`
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

func UserRolesFromDatabase(dbRoles []db.AclUserRole) []UserRole {
	var out []UserRole
	for _, r := range dbRoles {
		facility := r.Facility
		if facility == acl.ScopedRoleGlobalFacility {
			facility = "ZHQ"
		}
		out = append(out, UserRole{
			Facility:  facility,
			Role:      r.Role,
			GrantedAt: r.GrantedAt,
		})
	}
	return out
}

func UsersFromDatabase(users []db.GetCombinedUserRow, canSeeSensitiveFields bool) []User {
	var out []User
	for _, user := range users {
		out = append(out, UserFromDatabase(user, canSeeSensitiveFields, nil))
	}
	return out
}

func UserFromDatabase(user db.GetCombinedUserRow, canSeeSensitiveFields bool, roles []UserRole) User {
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
		Roles:        roles,
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
