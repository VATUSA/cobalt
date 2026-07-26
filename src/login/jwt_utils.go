package login

import (
	"vatusa-cobalt/acl"
	"vatusa-cobalt/auth"
	"vatusa-cobalt/db"
)

func CreateTokenForUser(user db.GetCombinedUserRow) (string, error) {
	permissionHandler := acl.GetPermissionHandlerCache().GetHandlerForCid(int(user.Cid))
	globalPermissions := ""
	facilityPermissions := ""
	isStaff := false
	if permissionHandler != nil {
		globalPermissions = permissionHandler.GetGlobalPermissionsString()
		facilityPermissions = permissionHandler.GetFacilityPermissionsString()
		isStaff = permissionHandler.IsStaff()
	}
	return auth.CreateToken(int(user.Cid), user.DisplayName, globalPermissions, facilityPermissions, isStaff, user.Facility)
}
