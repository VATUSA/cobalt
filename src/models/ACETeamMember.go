package models

import (
	"vatusa-cobalt/config"
	"vatusa-cobalt/db"
)

type ACETeamMember struct {
	CID         int    `json:"cid"`
	Name        string `json:"name"`
	Rating      int    `json:"rating"`
	RatingShort string `json:"rating_short"`
}

func ACETeamMembersFromDatabase(users []db.GetUsersByRoleRow) []ACETeamMember {
	members := make([]ACETeamMember, len(users))
	for i, u := range users {
		rating := int(u.ControllerRating)
		if u.InstructorRating > 0 {
			rating = int(u.InstructorRating)
		}
		members[i] = ACETeamMember{
			CID:         int(u.Cid),
			Name:        u.DisplayName,
			Rating:      int(u.ControllerRating),
			RatingShort: config.RatingShort(rating),
		}
	}
	return members
}
