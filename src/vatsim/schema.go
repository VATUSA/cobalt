package vatsim

type MemberData struct {
	Id               int     `json:"id"`
	NameFirst        string  `json:"name_first"`
	NameLast         string  `json:"name_last"`
	Email            string  `json:"email"`
	CountyState      string  `json:"countystate"`
	Country          string  `json:"country"`
	Rating           int     `json:"rating"`
	PilotRating      int     `json:"pilotrating"`
	MilitaryRating   int     `json:"militaryrating"`
	SuspensionDate   *string `json:"susp_date"`
	RegistrationDate string  `json:"reg_date"`
	RegionId         string  `json:"region_id"`
	DivisionId       string  `json:"division_id"`
	SubdivisionId    *string `json:"subdivision_id"`
	LastRatingChange *string `json:"lastratingchange"`
}

type DivisionMemberData struct {
	Items []MemberData `json:"items"`
	Count int          `json:"count"`
}
