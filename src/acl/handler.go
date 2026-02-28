package acl

import (
	"fmt"
	"time"
)

type permissionKey string

func createGlobalPermissionKey(object Object, action Action) permissionKey {
	return permissionKey(fmt.Sprintf("%s:%s", object, action))
}

func createFacilityPermissionKey(facility string, object Object, action Action) permissionKey {
	return permissionKey(fmt.Sprintf("%s:%s:%s", facility, object, action))
}

type PermissionHandler struct {
	globalPermissions   map[permissionKey]bool
	facilityPermissions map[permissionKey]bool
	loadTime            time.Time
}

func NewPermissionHandler(roles []ScopedRole) *PermissionHandler {
	ph := &PermissionHandler{
		globalPermissions:   make(map[permissionKey]bool),
		facilityPermissions: make(map[permissionKey]bool),
		loadTime:            time.Now(),
	}
	for _, role := range roles {
		globalPermissions := globalPermissionsForRole(role.Role)
		for _, f := range globalPermissions {
			ph.addGlobalPermission(f.Object, f.Action)
		}
		facilityPermissions := facilityPermissionsForRole(role.Role)
		for _, f := range facilityPermissions {
			if role.Facility == ScopedRoleGlobalFacility {
				ph.addGlobalPermission(f.Object, f.Action)
			} else {
				ph.addFacilityPermission(role.Facility, f.Object, f.Action)
			}
		}
	}
	return ph
}

func (ph *PermissionHandler) IsStale() bool {
	return ph.loadTime.Before(time.Now().Add(-time.Minute))
}

func (ph *PermissionHandler) addGlobalPermission(object Object, action Action) {
	key := createGlobalPermissionKey(object, action)
	ph.globalPermissions[key] = true
}

func (ph *PermissionHandler) addFacilityPermission(facility string, object Object, action Action) {
	key := createFacilityPermissionKey(facility, object, action)
	ph.facilityPermissions[key] = true
}

func (ph *PermissionHandler) HasGlobal(object Object, action Action) bool {
	saRes, ok := ph.globalPermissions[createGlobalPermissionKey(ObjectSuperAdmin, ActionUsage)]
	if ok {
		return saRes
	}
	res, ok := ph.globalPermissions[createGlobalPermissionKey(object, action)]
	if !ok {
		return false
	}
	return res
}

func (ph *PermissionHandler) HasFacility(facility string, object Object, action Action) bool {
	if ph.HasGlobal(object, action) {
		return true
	}
	res, ok := ph.facilityPermissions[createFacilityPermissionKey(facility, object, action)]
	if !ok {
		return false
	}
	return res
}
