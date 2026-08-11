package acl

import (
	"slices"
	"testing"
)

func roleInFacility(got map[string][]Role, facility string, role Role) bool {
	list, ok := got[facility]
	if !ok {
		return false
	}
	return slices.Contains(list, role)
}

func facilityKeys(got map[string][]Role) []string {
	keys := make([]string, 0, len(got))
	for key := range got {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

func TestGetAssignableRoles(t *testing.T) {
	t.Run("system admin can assign every listed role division-wide", func(t *testing.T) {
		ph := NewPermissionHandler([]ScopedRole{{Role: RoleSystemAdmin, Facility: ScopedRoleGlobalFacility}})
		got := ph.GetAssignableRoles()

		roles, ok := got["ZHQ"]
		if !ok {
			t.Fatalf("missing division-wide group, got facilities %v", facilityKeys(got))
		}
		for role := range RoleToPermissionObjectMap {
			if !slices.Contains(roles, role) {
				t.Errorf("system admin cannot assign %s", role)
			}
		}
		for _, role := range AutomaticRoles {
			if slices.Contains(roles, role) {
				t.Errorf("automatic role %s must not be assignable", role)
			}
		}
		if len(got) != 1 {
			t.Errorf("expected only the division-wide group, got facilities %v", facilityKeys(got))
		}
	})

	t.Run("atm assigns junior staff and training roles at their facility", func(t *testing.T) {
		ph := NewPermissionHandler([]ScopedRole{{Role: RoleAirTrafficManager, Facility: "ZNY"}})
		got := ph.GetAssignableRoles()

		for _, role := range []Role{
			RoleEventCoordinator,
			RoleFacilityEngineer,
			RoleWebMaintainer,
			RoleMentor,
			RoleFacilityAcademyEditor,
		} {
			if !roleInFacility(got, "ZNY", role) {
				t.Errorf("expected %s at ZNY", role)
			}
		}
		for _, role := range []Role{
			RoleAirTrafficManager,
			RoleDeputyAirTrafficManager,
			RoleTrainingAdministrator,
			RoleDivisionStaff,
			RoleInstructor,
		} {
			if roleInFacility(got, "ZNY", role) {
				t.Errorf("unexpected %s at ZNY", role)
			}
		}
		if len(got) != 1 {
			t.Errorf("expected only ZNY, got facilities %v", facilityKeys(got))
		}
	})

	t.Run("division staff assigns staff, senior staff and junior staff roles division-wide", func(t *testing.T) {
		ph := NewPermissionHandler([]ScopedRole{{Role: RoleDivisionStaff, Facility: ScopedRoleGlobalFacility}})
		got := ph.GetAssignableRoles()

		roles, ok := got["ZHQ"]
		if !ok {
			t.Fatalf("missing division-wide group, got facilities %v", facilityKeys(got))
		}
		for _, role := range []Role{
			RoleACETeam,
			RoleDivisionTechTeam,
			RoleSocialMediaTeam,
			RoleDivisionCommandCenter,
			RoleInstructor,
			RoleDivisionAcademyEditor,
			RoleEmailUser,
			RoleAirTrafficManager,
			RoleDeputyAirTrafficManager,
			RoleTrainingAdministrator,
			RoleEventCoordinator,
			RoleFacilityEngineer,
			RoleWebMaintainer,
		} {
			if !slices.Contains(roles, role) {
				t.Errorf("expected %s at ZHQ", role)
			}
		}
		for _, role := range []Role{
			RoleSystemAdmin,
			RoleSystemInternal,
			RoleSystemExternal,
			RoleDivisionManagement,
		} {
			if slices.Contains(roles, role) {
				t.Errorf("unexpected %s at ZHQ", role)
			}
		}
		if len(got) != 1 {
			t.Errorf("expected only the division-wide group, got facilities %v", facilityKeys(got))
		}
	})

	t.Run("roles at special facilities are not assignable", func(t *testing.T) {
		ph := NewPermissionHandler([]ScopedRole{{Role: RoleAirTrafficManager, Facility: "ZAE"}})
		got := ph.GetAssignableRoles()

		if len(got) != 0 {
			t.Errorf("expected no groups, got facilities %v", facilityKeys(got))
		}
	})
}
