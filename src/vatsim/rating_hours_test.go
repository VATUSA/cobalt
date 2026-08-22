package vatsim

import (
	"testing"
	"vatusa-cobalt/config"
)

// TestGetHoursTotalForRating covers the rating-to-stat-field mapping used to
// decide whether a controller has met the 50-hour requirement, including the
// controller-rating bucket which sums several stat fields together (C1-C3,
// I1-I3) rather than reading a single field.
func TestGetHoursTotalForRating(t *testing.T) {
	stats := &MemberStatistics{
		S1: 10, S2: 20, S3: 30,
		C1: 1, C2: 2, C3: 3,
		I1: 4, I2: 5, I3: 6,
	}

	cases := []struct {
		name   string
		rating int
		want   float64
	}{
		{"student 1", config.RatingStudent1, 10},
		{"student 2", config.RatingStudent2, 20},
		{"student 3", config.RatingStudent3, 30},
		{"controller bucket sums C1-C3 and I1-I3", config.RatingController1, 1 + 2 + 3 + 4 + 5 + 6},
		{"unmapped rating returns zero", config.RatingSupervisor, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := getHoursTotalForRating(stats, tc.rating); got != tc.want {
				t.Errorf("getHoursTotalForRating(%d) = %v, want %v", tc.rating, got, tc.want)
			}
		})
	}
}
