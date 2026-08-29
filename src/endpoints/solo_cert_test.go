package endpoints

import (
	"testing"
	"time"
)

func TestParseSoloCertExpires(t *testing.T) {
	today := time.Now().UTC()
	dateOnly := func(d time.Time) string {
		return d.Format(time.DateOnly)
	}

	cases := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{"today", dateOnly(today), false},
		{"plus_45_days", dateOnly(today.AddDate(0, 0, soloCertMaxExpiresDays)), false},
		{"plus_46_days", dateOnly(today.AddDate(0, 0, soloCertMaxExpiresDays+1)), true},
		{"yesterday", dateOnly(today.AddDate(0, 0, -1)), true},
		{"malformed", "not-a-date", true},
		{"wrong_format", "08/29/2026", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseSoloCertExpires(tc.raw)
			if (err != nil) != tc.wantErr {
				t.Errorf("parseSoloCertExpires(%q) error = %v, wantErr %v", tc.raw, err, tc.wantErr)
			}
		})
	}
}
