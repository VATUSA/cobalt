package config

import "testing"

func TestIsSpecialFacility(t *testing.T) {
	cases := []struct {
		name     string
		facility string
		want     bool
	}{
		{name: "academy", facility: FacilityAcademy, want: true},
		{name: "nonmember", facility: FacilityNonMember, want: true},
		{name: "inactive", facility: FacilityInactive, want: true},
		{name: "notexists", facility: FacilityNotExists, want: true},
		{name: "zhq", facility: "ZHQ", want: false},
		{name: "artcc", facility: "ZNY", want: false},
		{name: "empty", facility: "", want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsSpecialFacility(tc.facility); got != tc.want {
				t.Errorf("IsSpecialFacility(%q) = %v, want %v", tc.facility, got, tc.want)
			}
		})
	}
}
