package models

type User struct {
	CID          int           `json:"cid"`
	NetworkUser  *NetworkUser  `json:"network_user"`
	DivisionUser *DivisionUser `json:"division_user"`
}

type NetworkUser struct {
	CID            int     `json:"cid"`
	FirstName      *string `json:"first_name"`
	LastName       *string `json:"last_name"`
	Email          *string `json:"email"`
	Rating         int     `json:"rating"`
	Region         string  `json:"region"`
	Division       string  `json:"division"`
	SubDivision    string  `json:"subdivision"`
	PilotRating    int     `json:"pilot_rating"`
	MilitaryRating int     `json:"military_rating"`
}

type DivisionUser struct {
	DisplayName            string   `json:"display_name"`
	ControllerRating       int      `json:"controller_rating"`
	InstructorRating       *int     `json:"instructor_rating"`
	Facility               string   `json:"facility"`
	VisitingFacilities     []string `json:"visiting_facilities"`
	DiscordId              int64    `json:"discord_id"`
	LastPromotionTimestamp int64    `json:"last_promotion_timestamp"`
	LastTransferTimestamp  int64    `json:"last_transfer_timestamp"`
}
