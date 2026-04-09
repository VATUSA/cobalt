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
	// TODO: Determine if the user is a member of the division and assign RoleDivisionMember role
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
