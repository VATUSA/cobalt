package models

import "vatusa-cobalt/db"

type FacilityStaffMember struct {
	Role        string `json:"role"`
	Cid         int    `json:"cid"`
	DisplayName string `json:"display_name"`
}

type FacilityStaff struct {
	Staff []FacilityStaffMember `json:"staff"`
}

func FacilityStaffFromDatabase(rows []db.GetFacilityStaffRolesRow) FacilityStaff {
	staff := []FacilityStaffMember{}
	for _, row := range rows {
		staff = append(staff, FacilityStaffMember{
			Role:        row.Role,
			Cid:         int(row.Cid),
			DisplayName: row.DisplayName,
		})
	}
	return FacilityStaff{Staff: staff}
}
