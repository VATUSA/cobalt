package models

import (
	"time"
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
	Id        int       `json:"id"`
	Cid       int       `json:"cid"`
	Facility  string    `json:"facility"`
	Position  string    `json:"position"`
	Expires   time.Time `json:"expires"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func SoloCertFromDatabase(ent db.SoloCert) SoloCert {
	return SoloCert{
		Id:        int(ent.ID),
		Cid:       int(ent.Cid),
		Facility:  ent.Facility,
		Position:  ent.Position,
		Expires:   ent.Expires,
		CreatedAt: ent.CreatedAt,
		UpdatedAt: ent.UpdatedAt,
	}
}

func SoloCertsFromDatabase(ents []db.SoloCert) []SoloCert {
	items := make([]SoloCert, len(ents))
	for i, ent := range ents {
		items[i] = SoloCertFromDatabase(ent)
	}
	return items
}
