package acl

import (
	"context"
	"vatusa-cobalt/dbconn"
)

func GetUserScopedRoles(cid int) ([]ScopedRole, error) {
	ctx := context.Background()
	roles, err := dbconn.Queries().GetRolesForUser(ctx, int32(cid))
	if err != nil {
		return nil, err
	}
	output := []ScopedRole{
		{
			Facility: ScopedRoleGlobalFacility,
			Role:     RoleAuthenticatedUser,
		},
	}
	user, err := dbconn.GetCombinedUserByCID(cid)
	if err == nil && user != nil && user.DivisionID == "USA" {
		output = append(output, ScopedRole{
			Facility: ScopedRoleGlobalFacility,
			Role:     RoleDivisionMember,
		})
	}
	for _, role := range roles {
		output = append(output, ScopedRole{
			Role:     Role(role.Role),
			Facility: role.Facility,
		})
	}
	return output, nil
}

func GetActorScopedRoles(actorId int) ([]ScopedRole, error) {

	ctx := context.Background()
	roles, err := dbconn.Queries().GetRolesForActor(ctx, int32(actorId))
	if err != nil {
		return nil, err
	}
	var output []ScopedRole
	for _, role := range roles {
		output = append(output, ScopedRole{
			Role:     Role(role.Role),
			Facility: role.Facility,
		})
	}
	return output, nil
}
