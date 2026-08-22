package models

import "vatusa-cobalt/db"

type FacilityTitleRequest struct {
	Title string `json:"title"`
	Code  string `json:"code"`
	Tier  string `json:"tier"`
}

type AssignFacilityTitleRequest struct {
	TitleId int64 `json:"title_id"`
}

type FacilityTitle struct {
	Id        int64  `json:"id"`
	Facility  string `json:"facility"`
	Title     string `json:"title"`
	Code      string `json:"code"`
	Tier      string `json:"tier"`
	CreatedAt int64  `json:"created_at"`
}

func FacilityTitleFromDatabase(ent db.FacilityTitle) FacilityTitle {
	return FacilityTitle{
		Id:        ent.ID,
		Facility:  ent.Facility,
		Title:     ent.Title,
		Code:      ent.Code,
		Tier:      ent.Tier,
		CreatedAt: ent.CreatedAt,
	}
}

func FacilityTitlesFromDatabase(ents []db.FacilityTitle) []FacilityTitle {
	titles := make([]FacilityTitle, len(ents))
	for i, ent := range ents {
		titles[i] = FacilityTitleFromDatabase(ent)
	}
	return titles
}

type UserFacilityTitle struct {
	Id         int64  `json:"id"`
	Facility   string `json:"facility"`
	Title      string `json:"title"`
	Code       string `json:"code"`
	Tier       string `json:"tier"`
	GrantorCid int32  `json:"grantor_cid"`
	GrantedAt  int64  `json:"granted_at"`
}

func UserFacilityTitleFromDatabase(ent db.GetUserTitlesByFacilityRow) UserFacilityTitle {
	return UserFacilityTitle{
		Id:         ent.ID,
		Facility:   ent.Facility,
		Title:      ent.Title,
		Code:       ent.Code,
		Tier:       ent.Tier,
		GrantorCid: ent.GrantorCid,
		GrantedAt:  ent.GrantedAt,
	}
}

func UserFacilityTitlesFromDatabase(ents []db.GetUserTitlesByFacilityRow) []UserFacilityTitle {
	titles := make([]UserFacilityTitle, len(ents))
	for i, ent := range ents {
		titles[i] = UserFacilityTitleFromDatabase(ent)
	}
	return titles
}
