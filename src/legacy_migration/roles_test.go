package legacy_migration

import (
	"testing"
	"vatusa-cobalt/acl"
	"vatusa-cobalt/models"
)

func TestLegacyToModernRoles_DirectMapping(t *testing.T) {
	roles := []models.LegacyRole{
		{Role: "ATM", Facility: "ZDV"},
		{Role: "MTR", Facility: "ZLA"},
	}

	got := LegacyToModernRoles(roles)

	want := []acl.ScopedRole{
		{Role: acl.RoleAirTrafficManager, Facility: "ZDV"},
		{Role: acl.RoleMentor, Facility: "ZLA"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d roles, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("role %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

// TestLegacyToModernRoles_GlobalRoleFacilityIsForced covers that a directly
// mapped role which is global (e.g. CBT -> RoleDivisionAcademyEditor) is
// always scoped to the global facility, regardless of what facility the
// legacy row carried.
func TestLegacyToModernRoles_GlobalRoleFacilityIsForced(t *testing.T) {
	roles := []models.LegacyRole{
		{Role: "CBT", Facility: "ZDV"},
	}
	got := LegacyToModernRoles(roles)
	if len(got) != 1 {
		t.Fatalf("got %d roles, want 1: %+v", len(got), got)
	}
	if got[0].Facility != acl.ScopedRoleGlobalFacility {
		t.Errorf("expected a global role's facility to be forced to %q, got %q", acl.ScopedRoleGlobalFacility, got[0].Facility)
	}
}

// TestLegacyToModernRoles_DivisionStaffPrefix covers the fallback path for
// legacy "USn" roles that aren't directly mapped: US1-4 grant division
// management on top of the baseline division staff role, US0/US6 also grant
// system admin, and USWT is excluded from this fallback since it's handled
// by the direct map instead.
func TestLegacyToModernRoles_DivisionStaffPrefix(t *testing.T) {
	cases := []struct {
		legacyRole string
		want       []acl.Role
	}{
		{"US0", []acl.Role{acl.RoleDivisionStaff, acl.RoleSystemAdmin}},
		{"US1", []acl.Role{acl.RoleDivisionStaff, acl.RoleDivisionManagement}},
		{"US5", []acl.Role{acl.RoleDivisionStaff}},
		{"US6", []acl.Role{acl.RoleDivisionStaff, acl.RoleSystemAdmin}},
		{"US7", []acl.Role{acl.RoleDivisionStaff}},
	}

	for _, tc := range cases {
		t.Run(tc.legacyRole, func(t *testing.T) {
			got := LegacyToModernRoles([]models.LegacyRole{{Role: tc.legacyRole, Facility: "ZDV"}})
			if len(got) != len(tc.want) {
				t.Fatalf("got %d roles %+v, want %d roles matching %+v", len(got), got, len(tc.want), tc.want)
			}
			for i, r := range tc.want {
				if got[i].Role != r {
					t.Errorf("role %d = %s, want %s", i, got[i].Role, r)
				}
				if got[i].Facility != acl.ScopedRoleGlobalFacility {
					t.Errorf("expected role %d to be globally scoped, got facility %q", i, got[i].Facility)
				}
			}
		})
	}
}

func TestLegacyToModernRoles_USWTExcludedFromPrefixFallback(t *testing.T) {
	// USWT is handled by the direct map (-> RoleDivisionTechTeam) and must
	// not also fall through to the "USn" division-staff heuristic, which
	// checks the prefix "US" without excluding USWT explicitly in the map
	// lookup path.
	got := LegacyToModernRoles([]models.LegacyRole{{Role: "USWT", Facility: "ZDV"}})
	for _, r := range got {
		if r.Role == acl.RoleDivisionStaff {
			t.Errorf("did not expect USWT to also produce RoleDivisionStaff via the prefix fallback: %+v", got)
		}
	}
}

func TestLegacyToModernRoles_UnknownRoleIgnored(t *testing.T) {
	got := LegacyToModernRoles([]models.LegacyRole{{Role: "NOTAROLE", Facility: "ZDV"}})
	if len(got) != 0 {
		t.Errorf("expected an unrecognized, non-US-prefixed role to be dropped, got %+v", got)
	}
}

func TestLegacyToModernRoles_EmptyInput(t *testing.T) {
	got := LegacyToModernRoles(nil)
	if len(got) != 0 {
		t.Errorf("expected no roles for empty input, got %+v", got)
	}
}

func TestDetermineDropAddRoles(t *testing.T) {
	existing := []acl.ScopedRole{
		{Role: acl.RoleAirTrafficManager, Facility: "ZDV"},
		{Role: acl.RoleMentor, Facility: "ZDV"},
	}
	target := []acl.ScopedRole{
		{Role: acl.RoleMentor, Facility: "ZDV"},
		{Role: acl.RoleEventCoordinator, Facility: "ZDV"},
	}

	add, drop := DetermineDropAddRoles(existing, target)

	if len(add) != 1 || add[0].Role != acl.RoleEventCoordinator {
		t.Errorf("add = %+v, want exactly [RoleEventCoordinator]", add)
	}
	if len(drop) != 1 || drop[0].Role != acl.RoleAirTrafficManager {
		t.Errorf("drop = %+v, want exactly [RoleAirTrafficManager]", drop)
	}
}

func TestDetermineDropAddRoles_SameFacilityDifferentiatesRoles(t *testing.T) {
	// The same Role at two different facilities must be treated as distinct
	// grants: dropping ZDV's grant should not be masked by ZLA's still
	// existing, and vice versa.
	existing := []acl.ScopedRole{
		{Role: acl.RoleMentor, Facility: "ZDV"},
	}
	target := []acl.ScopedRole{
		{Role: acl.RoleMentor, Facility: "ZLA"},
	}

	add, drop := DetermineDropAddRoles(existing, target)

	if len(add) != 1 || add[0].Facility != "ZLA" {
		t.Errorf("add = %+v, want exactly [{RoleMentor ZLA}]", add)
	}
	if len(drop) != 1 || drop[0].Facility != "ZDV" {
		t.Errorf("drop = %+v, want exactly [{RoleMentor ZDV}]", drop)
	}
}

func TestDetermineDropAddRoles_NoChanges(t *testing.T) {
	roles := []acl.ScopedRole{
		{Role: acl.RoleAirTrafficManager, Facility: "ZDV"},
	}
	add, drop := DetermineDropAddRoles(roles, roles)
	if len(add) != 0 || len(drop) != 0 {
		t.Errorf("expected no changes for identical role sets, got add=%+v drop=%+v", add, drop)
	}
}

func TestDetermineDropAddRoles_EmptyExisting(t *testing.T) {
	target := []acl.ScopedRole{{Role: acl.RoleMentor, Facility: "ZDV"}}
	add, drop := DetermineDropAddRoles(nil, target)
	if len(drop) != 0 {
		t.Errorf("expected no drops with no existing roles, got %+v", drop)
	}
	if len(add) != 1 {
		t.Errorf("expected the entire target set to be added, got %+v", add)
	}
}

func TestDetermineDropAddRoles_EmptyTarget(t *testing.T) {
	existing := []acl.ScopedRole{{Role: acl.RoleMentor, Facility: "ZDV"}}
	add, drop := DetermineDropAddRoles(existing, nil)
	if len(add) != 0 {
		t.Errorf("expected nothing added with an empty target, got %+v", add)
	}
	if len(drop) != 1 {
		t.Errorf("expected the entire existing set to be dropped, got %+v", drop)
	}
}
