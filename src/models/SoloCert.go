package models

import (
	"time"
	"vatusa-cobalt/config"
	"vatusa-cobalt/db"
)

type SoloCertRequest struct {
	Cid      int    `json:"cid"`
	Position string `json:"position"`
	Expires  string `json:"expires"`
}

type SoloCertUpdateRequest struct {
	Position string `json:"position"`
	Expires  string `json:"expires"`
}

type SoloCert struct {
	Id           int    `json:"id"`
	Cid          int    `json:"cid"`
	Facility     string `json:"facility"`
	Position     string `json:"position"`
	Expires      string `json:"expires"`
	CreatedByCid int    `json:"created_by_cid"`
	UpdatedByCid int    `json:"updated_by_cid"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

func SoloCertFromDatabase(ent db.SoloCert) SoloCert {
	return SoloCert{
		Id:           int(ent.ID),
		Cid:          int(ent.Cid),
		Facility:     ent.Facility,
		Position:     ent.Position,
		Expires:      ent.Expires.Format(time.DateOnly),
		CreatedByCid: int(ent.CreatedByCid),
		UpdatedByCid: int(ent.UpdatedByCid),
		CreatedAt:    ent.CreatedAt.Format(config.TimestampFormat),
		UpdatedAt:    ent.UpdatedAt.Format(config.TimestampFormat),
	}
}

func SoloCertsFromDatabase(ents []db.SoloCert) []SoloCert {
	items := make([]SoloCert, len(ents))
	for i, ent := range ents {
		items[i] = SoloCertFromDatabase(ent)
	}
	return items
}
