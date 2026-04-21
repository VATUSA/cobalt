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

type MemberStatistics struct {
	Id    int     `json:"id"`
	ATC   float64 `json:"atc"`
	Pilot float64 `json:"pilot"`
	S1    float64 `json:"s1"`
	S2    float64 `json:"s2"`
	S3    float64 `json:"s3"`
	C1    float64 `json:"c1"`
	C2    float64 `json:"c2"`
	C3    float64 `json:"c3"`
	I1    float64 `json:"i1"`
	I2    float64 `json:"i2"`
	I3    float64 `json:"i3"`
	SUP   float64 `json:"sup"`
	ADM   float64 `json:"adm"`
}
