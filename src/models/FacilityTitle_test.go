package models

import (
	"testing"
	"vatusa-cobalt/db"
)

func TestFacilityTitleFromDatabase(t *testing.T) {
	cases := []struct {
		name string
		ent  db.FacilityTitle
		want FacilityTitle
	}{
		{
			name: "full record",
			ent: db.FacilityTitle{
				ID:        7,
				Facility:  "ZNY",
				Title:     "Air Traffic Manager",
				Code:      "ATM",
				Tier:      "senior",
				CreatedAt: 12345,
			},
			want: FacilityTitle{
				Id:        7,
				Facility:  "ZNY",
				Title:     "Air Traffic Manager",
				Code:      "ATM",
				Tier:      "senior",
				CreatedAt: 12345,
			},
		},
		{
			name: "empty fields",
			ent: db.FacilityTitle{
				ID:        1,
				Facility:  "ZHQ",
				Title:     "Division Director",
				Code:      "US1",
				Tier:      "senior",
				CreatedAt: 0,
			},
			want: FacilityTitle{
				Id:        1,
				Facility:  "ZHQ",
				Title:     "Division Director",
				Code:      "US1",
				Tier:      "senior",
				CreatedAt: 0,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := FacilityTitleFromDatabase(tc.ent); got != tc.want {
				t.Errorf("FacilityTitleFromDatabase() = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestFacilityTitlesFromDatabase(t *testing.T) {
	cases := []struct {
		name string
		ents []db.FacilityTitle
		want []FacilityTitle
	}{
		{
			name: "empty input",
			ents: []db.FacilityTitle{},
			want: []FacilityTitle{},
		},
		{
			name: "multiple titles",
			ents: []db.FacilityTitle{
				{ID: 1, Facility: "ZNY", Title: "Air Traffic Manager", Code: "ATM", Tier: "senior", CreatedAt: 1},
				{ID: 2, Facility: "ZNY", Title: "Mentor", Code: "MTR", Tier: "junior", CreatedAt: 2},
			},
			want: []FacilityTitle{
				{Id: 1, Facility: "ZNY", Title: "Air Traffic Manager", Code: "ATM", Tier: "senior", CreatedAt: 1},
				{Id: 2, Facility: "ZNY", Title: "Mentor", Code: "MTR", Tier: "junior", CreatedAt: 2},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FacilityTitlesFromDatabase(tc.ents)
			if len(got) != len(tc.want) {
				t.Fatalf("got %d titles, want %d", len(got), len(tc.want))
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("title[%d] = %+v, want %+v", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestUserFacilityTitleFromDatabase(t *testing.T) {
	cases := []struct {
		name string
		ent  db.GetUserTitlesByFacilityRow
		want UserFacilityTitle
	}{
		{
			name: "full record",
			ent: db.GetUserTitlesByFacilityRow{
				ID:         5,
				Facility:   "ZNY",
				Title:      "Instructor",
				Code:       "INS",
				Tier:       "junior",
				GrantorCid: 123456,
				GrantedAt:  999,
			},
			want: UserFacilityTitle{
				Id:         5,
				Facility:   "ZNY",
				Title:      "Instructor",
				Code:       "INS",
				Tier:       "junior",
				GrantorCid: 123456,
				GrantedAt:  999,
			},
		},
		{
			name: "zero grantor",
			ent: db.GetUserTitlesByFacilityRow{
				ID:         3,
				Facility:   "ZHQ",
				Title:      "Technical Manager",
				Code:       "US6",
				Tier:       "senior",
				GrantorCid: 0,
				GrantedAt:  0,
			},
			want: UserFacilityTitle{
				Id:         3,
				Facility:   "ZHQ",
				Title:      "Technical Manager",
				Code:       "US6",
				Tier:       "senior",
				GrantorCid: 0,
				GrantedAt:  0,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := UserFacilityTitleFromDatabase(tc.ent); got != tc.want {
				t.Errorf("UserFacilityTitleFromDatabase() = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestUserFacilityTitlesFromDatabase(t *testing.T) {
	cases := []struct {
		name string
		ents []db.GetUserTitlesByFacilityRow
		want []UserFacilityTitle
	}{
		{
			name: "empty input",
			ents: []db.GetUserTitlesByFacilityRow{},
			want: []UserFacilityTitle{},
		},
		{
			name: "multiple titles",
			ents: []db.GetUserTitlesByFacilityRow{
				{ID: 1, Facility: "ZNY", Title: "Mentor", Code: "MTR", Tier: "junior", GrantorCid: 1, GrantedAt: 10},
				{ID: 2, Facility: "ZNY", Title: "Events Coordinator", Code: "EC", Tier: "junior", GrantorCid: 2, GrantedAt: 20},
			},
			want: []UserFacilityTitle{
				{Id: 1, Facility: "ZNY", Title: "Mentor", Code: "MTR", Tier: "junior", GrantorCid: 1, GrantedAt: 10},
				{Id: 2, Facility: "ZNY", Title: "Events Coordinator", Code: "EC", Tier: "junior", GrantorCid: 2, GrantedAt: 20},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := UserFacilityTitlesFromDatabase(tc.ents)
			if len(got) != len(tc.want) {
				t.Fatalf("got %d titles, want %d", len(got), len(tc.want))
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("title[%d] = %+v, want %+v", i, got[i], tc.want[i])
				}
			}
		})
	}
}
