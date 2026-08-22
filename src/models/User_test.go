package models

import (
	"database/sql"
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

func fullUserRow() db.GetCombinedUserRow {
	return db.GetCombinedUserRow{
		Cid:                654321,
		NameFirst:          "Jane",
		NameLast:           "Doe",
		Email:              "jane@example.com",
		Rating:             5,
		Pilotrating:        1,
		Militaryrating:     0,
		RegionID:           "AMAS",
		DivisionID:         "USA",
		SubdivisionID:      sql.NullString{String: "SUB", Valid: true},
		DisplayName:        "Jane D.",
		ControllerRating:   5,
		InstructorRating:   0,
		Facility:           "ZDV",
		VisitingFacilities: sql.NullString{String: "ZLA,ZOA", Valid: true},
		DiscordID:          "12345",
		LastPromotionTime:  sql.NullTime{Valid: false},
		LastTransferTime:   sql.NullTime{Valid: false},
	}
}

func TestUserFromDatabase_SensitiveFieldsHiddenByDefault(t *testing.T) {
	out := UserFromDatabase(fullUserRow(), false, nil)

	if out.NetworkUser.FirstName != nil {
		t.Error("expected FirstName to be hidden when canSeeSensitiveFields is false")
	}
	if out.NetworkUser.LastName != nil {
		t.Error("expected LastName to be hidden when canSeeSensitiveFields is false")
	}
	if out.NetworkUser.Email != nil {
		t.Error("expected Email to be hidden when canSeeSensitiveFields is false")
	}
}

func TestUserFromDatabase_SensitiveFieldsShownWhenPermitted(t *testing.T) {
	out := UserFromDatabase(fullUserRow(), true, nil)

	if out.NetworkUser.FirstName == nil || *out.NetworkUser.FirstName != "Jane" {
		t.Errorf("FirstName = %v, want Jane", out.NetworkUser.FirstName)
	}
	if out.NetworkUser.LastName == nil || *out.NetworkUser.LastName != "Doe" {
		t.Errorf("LastName = %v, want Doe", out.NetworkUser.LastName)
	}
	if out.NetworkUser.Email == nil || *out.NetworkUser.Email != "jane@example.com" {
		t.Errorf("Email = %v, want jane@example.com", out.NetworkUser.Email)
	}
}

func TestUserFromDatabase_NullableFieldsOmittedWhenInvalid(t *testing.T) {
	row := fullUserRow()
	row.SubdivisionID = sql.NullString{Valid: false}
	row.VisitingFacilities = sql.NullString{Valid: false}
	row.DiscordID = ""
	row.LastPromotionTime = sql.NullTime{Valid: false}
	row.LastTransferTime = sql.NullTime{Valid: false}

	out := UserFromDatabase(row, false, nil)

	if out.NetworkUser.SubDivision != nil {
		t.Errorf("expected nil SubDivision, got %v", *out.NetworkUser.SubDivision)
	}
	if len(out.DivisionUser.VisitingFacilities) != 0 {
		t.Errorf("expected an empty (not nil-panicking) VisitingFacilities slice, got %v", out.DivisionUser.VisitingFacilities)
	}
	if out.DivisionUser.DiscordId != nil {
		t.Errorf("expected nil DiscordId for an empty discord id, got %v", *out.DivisionUser.DiscordId)
	}
	if out.DivisionUser.LastPromotionTimestamp != nil {
		t.Error("expected nil LastPromotionTimestamp when not valid")
	}
	if out.DivisionUser.LastTransferTimestamp != nil {
		t.Error("expected nil LastTransferTimestamp when not valid")
	}
}

func TestUserFromDatabase_VisitingFacilitiesSplit(t *testing.T) {
	row := fullUserRow()
	row.VisitingFacilities = sql.NullString{String: "ZLA,ZOA,ZDV", Valid: true}

	out := UserFromDatabase(row, false, nil)

	want := []string{"ZLA", "ZOA", "ZDV"}
	if len(out.DivisionUser.VisitingFacilities) != len(want) {
		t.Fatalf("VisitingFacilities = %v, want %v", out.DivisionUser.VisitingFacilities, want)
	}
	for i := range want {
		if out.DivisionUser.VisitingFacilities[i] != want[i] {
			t.Errorf("VisitingFacilities[%d] = %q, want %q", i, out.DivisionUser.VisitingFacilities[i], want[i])
		}
	}
}

func TestUsersFromDatabase_MapsEachRow(t *testing.T) {
	rows := []db.GetCombinedUserRow{fullUserRow(), fullUserRow()}
	rows[1].Cid = 999

	out := UsersFromDatabase(rows, false)

	if len(out) != 2 {
		t.Fatalf("got %d users, want 2", len(out))
	}
	if out[0].CID != 654321 || out[1].CID != 999 {
		t.Errorf("CIDs = [%d, %d], want [654321, 999]", out[0].CID, out[1].CID)
	}
}
