package endpoints

import (
	"strings"
	"testing"
	"vatusa-cobalt/acl"
)

func TestValidateRoleFacility(t *testing.T) {
	cases := []struct {
		name     string
		role     acl.Role
		facility string
		wantErr  string
	}{
		{name: "unknown role", role: "not_a_role", facility: "ZNY", wantErr: "unknown role"},
		{name: "empty role", role: "", facility: "ZNY", wantErr: "unknown role"},

		{name: "atm at academy", role: acl.RoleAirTrafficManager, facility: "ZAE", wantErr: "this facility"},
		{name: "atm at nonmember", role: acl.RoleAirTrafficManager, facility: "ZZN", wantErr: "this facility"},
		{name: "atm at inactive", role: acl.RoleAirTrafficManager, facility: "ZZI", wantErr: "this facility"},
		{name: "atm at notexists", role: acl.RoleAirTrafficManager, facility: "ZZZ", wantErr: "this facility"},

		{name: "auto role at facility", role: acl.RoleAuthenticatedUser, facility: "ZNY", wantErr: "automatic"},
		{name: "auto role at zhq", role: acl.RoleDivisionMember, facility: "ZHQ", wantErr: "automatic"},

		{name: "division_staff at facility", role: acl.RoleDivisionStaff, facility: "ZNY", wantErr: "not a facility role"},
		{name: "division_management at facility", role: acl.RoleDivisionManagement, facility: "ZSE", wantErr: "not a facility role"},
		{name: "ace_team at facility", role: acl.RoleACETeam, facility: "ZLA", wantErr: "not a facility role"},

		{name: "atm at zhq", role: acl.RoleAirTrafficManager, facility: "ZHQ", wantErr: "not a global role"},
		{name: "datm at zhq", role: acl.RoleDeputyAirTrafficManager, facility: "ZHQ", wantErr: "not a global role"},
		{name: "ta at zhq", role: acl.RoleTrainingAdministrator, facility: "ZHQ", wantErr: "not a global role"},
		{name: "ec at zhq", role: acl.RoleEventCoordinator, facility: "ZHQ", wantErr: "not a global role"},
		{name: "instructor at zhq", role: acl.RoleInstructor, facility: "ZHQ", wantErr: "not a global role"},
		{name: "mentor at zhq", role: acl.RoleMentor, facility: "ZHQ", wantErr: "not a global role"},

		{name: "atm at facility", role: acl.RoleAirTrafficManager, facility: "ZNY", wantErr: ""},
		{name: "datm at facility", role: acl.RoleDeputyAirTrafficManager, facility: "ZLA", wantErr: ""},
		{name: "ta at facility", role: acl.RoleTrainingAdministrator, facility: "ZSE", wantErr: ""},
		{name: "ec at facility", role: acl.RoleEventCoordinator, facility: "ZAU", wantErr: ""},
		{name: "fe at facility", role: acl.RoleFacilityEngineer, facility: "ZHU", wantErr: ""},
		{name: "wm at facility", role: acl.RoleWebMaintainer, facility: "ZTL", wantErr: ""},
		{name: "instructor at facility", role: acl.RoleInstructor, facility: "ZOB", wantErr: ""},
		{name: "mentor at facility", role: acl.RoleMentor, facility: "ZKC", wantErr: ""},

		{name: "division_staff at zhq", role: acl.RoleDivisionStaff, facility: "ZHQ", wantErr: ""},
		{name: "division_management at zhq", role: acl.RoleDivisionManagement, facility: "ZHQ", wantErr: ""},
		{name: "ace_team at zhq", role: acl.RoleACETeam, facility: "ZHQ", wantErr: ""},
		{name: "division_tech_team at zhq", role: acl.RoleDivisionTechTeam, facility: "ZHQ", wantErr: ""},

		{name: "email_user at facility", role: acl.RoleEmailUser, facility: "ZNY", wantErr: ""},
		{name: "email_user at zhq", role: acl.RoleEmailUser, facility: "ZHQ", wantErr: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := validateRoleFacility(tc.role, tc.facility)
			if tc.wantErr == "" && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tc.wantErr != "" && err == nil {
				t.Errorf("expected error containing %q, got nil", tc.wantErr)
			}
			if tc.wantErr != "" && err != nil && !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error = %q, want substring %q", err.Error(), tc.wantErr)
			}
		})
	}
}
