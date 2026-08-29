package acl

import "testing"

// TestPermissionHandler_HasGlobal covers the core grant/deny logic that every
// endpoint permission check depends on.
func TestPermissionHandler_HasGlobal(t *testing.T) {
	ph := NewPermissionHandler([]ScopedRole{
		{Role: RoleDivisionManagement, Facility: ScopedRoleGlobalFacility},
	})

	if !ph.HasGlobal(ObjectDivisionManagementRole, ActionWrite) {
		t.Error("expected division management to have write on ObjectDivisionManagementRole")
	}
	if ph.HasGlobal(ObjectDivisionManagementRole, ActionRead) {
		t.Error("did not expect read permission that was never granted")
	}
	if ph.HasGlobal(ObjectSystemRole, ActionWrite) {
		t.Error("did not expect a permission from an unrelated role")
	}
}

// TestPermissionHandler_SuperAdminOverridesEverything covers the escape
// hatch: a superadmin grant must short-circuit every other global check,
// including permissions that were never explicitly granted.
func TestPermissionHandler_SuperAdminOverridesEverything(t *testing.T) {
	ph := NewPermissionHandler([]ScopedRole{
		{Role: RoleSystemAdmin, Facility: ScopedRoleGlobalFacility},
	})

	if !ph.HasGlobal(ObjectEventApproval, ActionWrite) {
		t.Error("expected superadmin to be granted an arbitrary object/action pair")
	}
	if !ph.HasGlobal(ObjectSuperAdmin, ActionUsage) {
		t.Error("expected superadmin usage grant itself to report true")
	}
}

// TestPermissionHandler_HasAny covers the fast-path check used to reject a
// caller with no grant of an object/action at any scope before looking up a
// specific target — without it, a lookup-then-403 flow can leak via a
// 404-vs-403 split whether a target exists to a caller who could never have
// permission on it regardless.
func TestPermissionHandler_HasAny(t *testing.T) {
	ph := NewPermissionHandler([]ScopedRole{
		{Role: RoleAirTrafficManager, Facility: "ZDV"},
	})

	if !ph.HasAny(ObjectEvent, ActionWrite) {
		t.Error("expected HasAny to report true for a facility-scoped grant")
	}
	if ph.HasAny(ObjectSystemRole, ActionWrite) {
		t.Error("did not expect HasAny to report true for an ungranted object")
	}

	empty := NewPermissionHandler([]ScopedRole{})
	if empty.HasAny(ObjectEvent, ActionWrite) {
		t.Error("did not expect HasAny to report true for a caller with no roles")
	}

	global := NewPermissionHandler([]ScopedRole{
		{Role: RoleDivisionManagement, Facility: ScopedRoleGlobalFacility},
	})
	if !global.HasAny(ObjectDivisionManagementRole, ActionWrite) {
		t.Error("expected HasAny to report true for a global grant")
	}
}

// TestPermissionHandler_HasFacility covers facility-scoped grants, and that
// a global grant of the same object/action still satisfies a facility check
// (a divison-wide permission should not need to be duplicated per facility).
func TestPermissionHandler_HasFacility(t *testing.T) {
	ph := NewPermissionHandler([]ScopedRole{
		{Role: RoleAirTrafficManager, Facility: "ZDV"},
	})

	if !ph.HasFacility("ZDV", ObjectEvent, ActionWrite) {
		t.Error("expected facility-scoped grant to apply to its own facility")
	}
	if ph.HasFacility("ZLA", ObjectEvent, ActionWrite) {
		t.Error("did not expect a facility-scoped grant to leak to another facility")
	}
}

func TestPermissionHandler_GlobalRoleAssignmentGrantsEveryFacility(t *testing.T) {
	// RoleAirTrafficManager normally grants ObjectFacilityJuniorStaffRole
	// scoped per-facility. Assigning it with the global facility ("*")
	// should promote that grant to global instead, per NewPermissionHandler.
	ph := NewPermissionHandler([]ScopedRole{
		{Role: RoleAirTrafficManager, Facility: ScopedRoleGlobalFacility},
	})

	if !ph.HasFacility("ZDV", ObjectFacilityJuniorStaffRole, ActionWrite) {
		t.Error("expected a globally-scoped role assignment to grant its facility permissions everywhere")
	}
	if !ph.HasFacility("ANY", ObjectFacilityJuniorStaffRole, ActionWrite) {
		t.Error("expected the global grant to apply to any facility code")
	}
}

func TestPermissionHandler_UnknownRoleGrantsNothing(t *testing.T) {
	ph := NewPermissionHandler([]ScopedRole{
		{Role: Role("not_a_real_role"), Facility: ScopedRoleGlobalFacility},
	})

	if ph.HasGlobal(ObjectSystemRole, ActionWrite) {
		t.Error("did not expect any permission from an unrecognized role")
	}
	if !ph.IsStaff() {
		t.Error("an unrecognized role should still count as staff since it isn't in AutomaticRoles")
	}
}

func TestPermissionHandler_IsStaff(t *testing.T) {
	cases := []struct {
		name  string
		roles []ScopedRole
		want  bool
	}{
		{
			name:  "only automatic roles",
			roles: []ScopedRole{{Role: RoleAuthenticatedUser, Facility: ScopedRoleGlobalFacility}},
			want:  false,
		},
		{
			name:  "no roles at all",
			roles: nil,
			want:  false,
		},
		{
			name: "mix of automatic and staff role",
			roles: []ScopedRole{
				{Role: RoleAuthenticatedUser, Facility: ScopedRoleGlobalFacility},
				{Role: RoleMentor, Facility: "ZDV"},
			},
			want: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ph := NewPermissionHandler(tc.roles)
			if got := ph.IsStaff(); got != tc.want {
				t.Errorf("IsStaff() = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestPermissionHandler_MaskedObjectsExcludedFromListings covers that role
// objects used to grant *other* roles (e.g. ObjectSystemRole) never show up
// in the permission listings returned to clients, even though they're used
// internally by HasGlobal/HasFacility.
func TestPermissionHandler_MaskedObjectsExcludedFromListings(t *testing.T) {
	ph := NewPermissionHandler([]ScopedRole{
		{Role: RoleSystemAdmin, Facility: ScopedRoleGlobalFacility},
		{Role: RoleAirTrafficManager, Facility: "ZDV"},
	})

	for _, perm := range ph.GetGlobalPermissions() {
		if perm.Object == ObjectSystemRole || perm.Object == ObjectDivisionManagementRole {
			t.Errorf("expected masked object %s to be excluded from GetGlobalPermissions", perm.Object)
		}
	}
	for _, perm := range ph.GetFacilityPermissions() {
		if perm.Object == ObjectFacilityTrainingRole {
			t.Errorf("expected masked object %s to be excluded from GetFacilityPermissions", perm.Object)
		}
	}

	// But the underlying grant must still function for HasGlobal/HasFacility.
	if !ph.HasGlobal(ObjectSystemRole, ActionWrite) {
		t.Error("masking from listings should not affect the actual permission check")
	}

	if got := ph.GetGlobalPermissionsString(); got == "" {
		t.Error("expected at least one non-masked permission in the string form")
	}
}

func TestPermissionHandler_NoRolesGrantsNothing(t *testing.T) {
	ph := NewPermissionHandler(nil)

	if ph.HasGlobal(ObjectEvent, ActionRead) {
		t.Error("expected no permissions with no roles assigned")
	}
	if ph.HasFacility("ZDV", ObjectEvent, ActionRead) {
		t.Error("expected no facility permissions with no roles assigned")
	}
	if ph.IsStaff() {
		t.Error("expected not staff with no roles assigned")
	}
}
