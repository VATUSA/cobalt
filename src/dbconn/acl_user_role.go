package dbconn

import (
	"context"
	"vatusa-cobalt/db"
)

func GetFacilityStaffRoles(facility string, roles []string) ([]db.GetFacilityStaffRolesRow, error) {
	ctx := context.Background()
	params := db.GetFacilityStaffRolesParams{
		Facility: facility,
		Roles:    roles,
	}
	return Queries().GetFacilityStaffRoles(ctx, params)
}
