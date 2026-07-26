package models

import "vatusa-cobalt/db"

type V3ApiKey struct {
	Id        int64   `json:"id"`
	Code      string  `json:"code"`
	Testing   bool    `json:"testing"`
	Facility  *string `json:"facility"`
	Notes     *string `json:"notes"`
	CreatedAt int64   `json:"created_at"`
	UpdatedAt *int64  `json:"updated_at"`
}

func V3ApiKeysFromDatabase(rows []db.V3ApiKey) []V3ApiKey {
	keys := make([]V3ApiKey, 0, len(rows))
	for _, row := range rows {
		key := V3ApiKey{
			Id:        row.ID,
			Code:      row.Code,
			Testing:   row.Testing,
			CreatedAt: row.CreatedAt,
		}
		if row.Facility.Valid {
			key.Facility = &row.Facility.String
		}
		if row.Notes.Valid {
			key.Notes = &row.Notes.String
		}
		if row.UpdatedAt.Valid {
			key.UpdatedAt = &row.UpdatedAt.Int64
		}
		keys = append(keys, key)
	}
	return keys
}
