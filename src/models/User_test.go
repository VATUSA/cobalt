package models

import (
	"testing"
	"vatusa-cobalt/acl"
	"vatusa-cobalt/db"
)

func TestUserRolesFromDatabase(t *testing.T) {
	cases := []struct {
		name  string
		roles []db.AclUserRole
		want  []UserRole
	}{
		{
			name:  "empty input",
			roles: []db.AclUserRole{},
			want:  []UserRole{},
		},
		{
			name: "global role facility maps to ZHQ",
			roles: []db.AclUserRole{
				{Facility: acl.ScopedRoleGlobalFacility, Role: string(acl.RoleDivisionStaff), GrantedAt: 123},
			},
			want: []UserRole{
				{Facility: "ZHQ", Role: string(acl.RoleDivisionStaff), GrantedAt: 123},
			},
		},
		{
			name: "facility role passes through unchanged",
			roles: []db.AclUserRole{
				{Facility: "ZNY", Role: string(acl.RoleAirTrafficManager), GrantedAt: 456},
			},
			want: []UserRole{
				{Facility: "ZNY", Role: string(acl.RoleAirTrafficManager), GrantedAt: 456},
			},
		},
		{
			name: "mixed global and facility roles",
			roles: []db.AclUserRole{
				{Facility: acl.ScopedRoleGlobalFacility, Role: string(acl.RoleDivisionManagement), GrantedAt: 1},
				{Facility: "ZSE", Role: string(acl.RoleDeputyAirTrafficManager), GrantedAt: 2},
			},
			want: []UserRole{
				{Facility: "ZHQ", Role: string(acl.RoleDivisionManagement), GrantedAt: 1},
				{Facility: "ZSE", Role: string(acl.RoleDeputyAirTrafficManager), GrantedAt: 2},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := UserRolesFromDatabase(tc.roles)
			if len(got) != len(tc.want) {
				t.Fatalf("got %d roles, want %d", len(got), len(tc.want))
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("role[%d] = %+v, want %+v", i, got[i], tc.want[i])
				}
			}
		})
	}
}
