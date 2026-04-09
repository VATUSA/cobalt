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
	if permissionHandler != nil {
		globalPermissions = permissionHandler.GetGlobalPermissionsString()
		facilityPermissions = permissionHandler.GetFacilityPermissionsString()
	}
	return auth.CreateToken(int(user.Cid), user.DisplayName, globalPermissions, facilityPermissions)
}
