package acl

import "testing"

func TestTitleTierToPermissionObjectMap(t *testing.T) {
	cases := []struct {
		name string
		tier TitleTier
		want Object
		ok   bool
	}{
		{name: "senior tier", tier: TitleTierSenior, want: ObjectFacilityTitleSeniorStaff, ok: true},
		{name: "junior tier", tier: TitleTierJunior, want: ObjectFacilityTitleJuniorStaff, ok: true},
		{name: "unknown tier", tier: TitleTier("trainee"), want: "", ok: false},
		{name: "empty tier", tier: TitleTier(""), want: "", ok: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := TitleTierToPermissionObjectMap[tc.tier]
			if ok != tc.ok {
				t.Errorf("TitleTierToPermissionObjectMap[%q] ok = %v, want %v", tc.tier, ok, tc.ok)
			}
			if got != tc.want {
				t.Errorf("TitleTierToPermissionObjectMap[%q] = %q, want %q", tc.tier, got, tc.want)
			}
		})
	}
}

func TestFacilityTitleManagementPermissions(t *testing.T) {
	ph := NewPermissionHandler([]ScopedRole{
		{Role: RoleDivisionManagement, Facility: ScopedRoleGlobalFacility},
	})

	cases := []struct {
		name   string
		object Object
		action Action
		want   bool
	}{
		{name: "title management write", object: ObjectFacilityTitleManagement, action: ActionWrite, want: true},
		{name: "senior staff write", object: ObjectFacilityTitleSeniorStaff, action: ActionWrite, want: true},
		{name: "junior staff write", object: ObjectFacilityTitleJuniorStaff, action: ActionWrite, want: true},
		{name: "title management read", object: ObjectFacilityTitleManagement, action: ActionRead, want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ph.HasGlobal(tc.object, tc.action); got != tc.want {
				t.Errorf("HasGlobal(%s:%s) = %v, want %v", tc.object, tc.action, got, tc.want)
			}
		})
	}
}

func TestFacilityTitleJuniorStaffPermissions(t *testing.T) {
	for _, role := range []Role{RoleAirTrafficManager, RoleDeputyAirTrafficManager, RoleTrainingAdministrator} {
		ph := NewPermissionHandler([]ScopedRole{
			{Role: role, Facility: "ZNY"},
		})

		t.Run(string(role)+" at home facility", func(t *testing.T) {
			if got := ph.HasFacility("ZNY", ObjectFacilityTitleJuniorStaff, ActionWrite); !got {
				t.Errorf("HasFacility(ZNY, junior, write) = false, want true for %s", role)
			}
		})
		t.Run(string(role)+" at other facility", func(t *testing.T) {
			if got := ph.HasFacility("ZSE", ObjectFacilityTitleJuniorStaff, ActionWrite); got {
				t.Errorf("HasFacility(ZSE, junior, write) = true, want false for %s", role)
			}
		})
		t.Run(string(role)+" no senior staff write", func(t *testing.T) {
			if got := ph.HasFacility("ZNY", ObjectFacilityTitleSeniorStaff, ActionWrite); got {
				t.Errorf("HasFacility(ZNY, senior, write) = true, want false for %s", role)
			}
		})
		t.Run(string(role)+" no title management", func(t *testing.T) {
			if got := ph.HasFacility("ZNY", ObjectFacilityTitleManagement, ActionWrite); got {
				t.Errorf("HasFacility(ZNY, title management, write) = true, want false for %s", role)
			}
		})
	}
}

func TestFacilityTitlePermissionsNotGrantedByDefault(t *testing.T) {
	cases := []struct {
		name  string
		roles []ScopedRole
	}{
		{name: "authenticated user", roles: []ScopedRole{{Role: RoleAuthenticatedUser, Facility: ScopedRoleGlobalFacility}}},
		{name: "division staff", roles: []ScopedRole{{Role: RoleDivisionStaff, Facility: ScopedRoleGlobalFacility}}},
		{name: "facility junior staff", roles: []ScopedRole{{Role: RoleEventCoordinator, Facility: "ZNY"}}},
		{name: "training staff", roles: []ScopedRole{{Role: RoleMentor, Facility: "ZNY"}}},
		{name: "empty roles", roles: []ScopedRole{}},
	}

	objects := []Object{ObjectFacilityTitleManagement, ObjectFacilityTitleSeniorStaff, ObjectFacilityTitleJuniorStaff}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ph := NewPermissionHandler(tc.roles)
			for _, object := range objects {
				if got := ph.HasGlobal(object, ActionWrite); got {
					t.Errorf("HasGlobal(%s:write) = true, want false for %s", object, tc.name)
				}
				if got := ph.HasFacility("ZNY", object, ActionWrite); got {
					t.Errorf("HasFacility(ZNY, %s:write) = true, want false for %s", object, tc.name)
				}
			}
		})
	}
}
