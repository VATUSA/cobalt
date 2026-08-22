package roster

import (
	"database/sql"
	"testing"
	"time"
	"vatusa-cobalt/config"
	"vatusa-cobalt/db"
)

// blockerByCode finds a single Blocker's Check function so tests can call it
// directly, without running through CheckUserBlockers/GetUserBlockers - two
// of the blockers in the table (missing_rating_hours, pending_transfer) hit
// a live database connection, so a full-table loop isn't safe in a unit test.
func blockerByCode(t *testing.T, code string) BlockerCheckFunc {
	t.Helper()
	for _, b := range Blockers {
		if b.Code == code {
			return b.Check
		}
	}
	t.Fatalf("no blocker registered with code %q", code)
	return nil
}

func baseUser() db.GetCombinedUserRow {
	return db.GetCombinedUserRow{
		Cid:              100001,
		RegionID:         config.RegionAmericas,
		DivisionID:       config.DivisionVATUSA,
		Facility:         "ZDV",
		Rating:           config.RatingController1,
		ControllerRating: config.RatingController1,
		LastCompetencyDate: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
	}
}

func TestBlocker_NotInDivision(t *testing.T) {
	check := blockerByCode(t, "not_in_division")

	cases := []struct {
		name   string
		mutate func(*db.GetCombinedUserRow)
		want   bool
	}{
		{"in division", func(u *db.GetCombinedUserRow) {}, false},
		{"wrong region", func(u *db.GetCombinedUserRow) { u.RegionID = "EUR" }, true},
		{"wrong division", func(u *db.GetCombinedUserRow) { u.DivisionID = "OTH" }, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			u := baseUser()
			tc.mutate(&u)
			if got := check(u); got != tc.want {
				t.Errorf("check() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestBlocker_NeedsExam(t *testing.T) {
	check := blockerByCode(t, "needs_exam")

	t.Run("no competency date on record blocks", func(t *testing.T) {
		u := baseUser()
		u.LastCompetencyDate = sql.NullTime{}
		if !check(u) {
			t.Error("expected block when LastCompetencyDate is not set")
		}
	})

	t.Run("recent competency does not block", func(t *testing.T) {
		u := baseUser()
		u.LastCompetencyDate = sql.NullTime{Time: time.Now().Add(-30 * 24 * time.Hour), Valid: true}
		if check(u) {
			t.Error("did not expect a block for a recent competency date")
		}
	})

	t.Run("stale competency blocks", func(t *testing.T) {
		u := baseUser()
		u.LastCompetencyDate = sql.NullTime{Time: time.Now().Add(-181 * 24 * time.Hour), Valid: true}
		if !check(u) {
			t.Error("expected a block for a competency date older than 180 days")
		}
	})
}

func TestBlocker_RecentTransfer(t *testing.T) {
	check := blockerByCode(t, "recent_transfer")

	t.Run("never transferred does not block", func(t *testing.T) {
		u := baseUser()
		if check(u) {
			t.Error("did not expect a block with no transfer on record")
		}
	})

	t.Run("recent transfer blocks", func(t *testing.T) {
		u := baseUser()
		u.LastTransferTime = sql.NullTime{Time: time.Now().Add(-10 * 24 * time.Hour), Valid: true}
		if !check(u) {
			t.Error("expected a block for a transfer within the last 90 days")
		}
	})

	t.Run("old transfer does not block", func(t *testing.T) {
		u := baseUser()
		u.LastTransferTime = sql.NullTime{Time: time.Now().Add(-91 * 24 * time.Hour), Valid: true}
		if check(u) {
			t.Error("did not expect a block for a transfer more than 90 days ago")
		}
	})

	t.Run("academy bypasses even a recent transfer", func(t *testing.T) {
		u := baseUser()
		u.Facility = config.FacilityAcademy
		u.LastTransferTime = sql.NullTime{Time: time.Now().Add(-1 * time.Hour), Valid: true}
		if check(u) {
			t.Error("expected academy members to bypass the recent_transfer block")
		}
	})
}

func TestBlocker_RecentPromotion(t *testing.T) {
	check := blockerByCode(t, "recent_promotion")

	u := baseUser()
	if check(u) {
		t.Error("did not expect a block with no promotion on record")
	}

	u.LastPromotionTime = sql.NullTime{Time: time.Now().Add(-89 * 24 * time.Hour), Valid: true}
	if !check(u) {
		t.Error("expected a block for a promotion within the last 90 days")
	}

	u.LastPromotionTime = sql.NullTime{Time: time.Now().Add(-91 * 24 * time.Hour), Valid: true}
	if check(u) {
		t.Error("did not expect a block for a promotion more than 90 days ago")
	}
}

func TestBlocker_InstructorRating(t *testing.T) {
	check := blockerByCode(t, "instructor_rating")

	cases := []struct {
		name   string
		mutate func(*db.GetCombinedUserRow)
		want   bool
	}{
		{"controller rating", func(u *db.GetCombinedUserRow) {}, false},
		{"instructor by Rating field", func(u *db.GetCombinedUserRow) { u.Rating = config.RatingInstructor }, true},
		{"senior instructor by Rating field", func(u *db.GetCombinedUserRow) { u.Rating = config.RatingSeniorInstructor }, true},
		{"instructor rating credential set", func(u *db.GetCombinedUserRow) { u.InstructorRating = 1 }, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			u := baseUser()
			tc.mutate(&u)
			if got := check(u); got != tc.want {
				t.Errorf("check() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestBlocker_InTrainingAcademy(t *testing.T) {
	check := blockerByCode(t, "in_training_academy")

	u := baseUser()
	if check(u) {
		t.Error("did not expect a block for a normal home facility")
	}
	u.Facility = config.FacilityAcademy
	if !check(u) {
		t.Error("expected a block for members homed in the training academy")
	}
}

func TestBlocker_IsInactive(t *testing.T) {
	check := blockerByCode(t, "is_inactive")

	cases := []struct {
		name   string
		rating int32
		want   bool
	}{
		{"active controller", config.RatingController1, false},
		{"inactive", config.RatingInactive, true},
		{"suspended", config.RatingSuspended, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			u := baseUser()
			u.Rating = tc.rating
			if got := check(u); got != tc.want {
				t.Errorf("check() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestBlocker_VisitLowRating(t *testing.T) {
	check := blockerByCode(t, "visit_low_rating")

	u := baseUser()
	u.Rating = config.RatingStudent2
	if !check(u) {
		t.Error("expected a block below S3")
	}
	u.Rating = config.RatingStudent3
	if check(u) {
		t.Error("did not expect a block at exactly S3")
	}
	u.Rating = config.RatingController1
	if check(u) {
		t.Error("did not expect a block above S3")
	}
}

func TestBlocker_RecentVisit(t *testing.T) {
	check := blockerByCode(t, "recent_visit")

	u := baseUser()
	if check(u) {
		t.Error("did not expect a block with no visit on record")
	}
	u.LastVisitTime = sql.NullTime{Time: time.Now().Add(-1 * 24 * time.Hour), Valid: true}
	if !check(u) {
		t.Error("expected a block for a visit within the last 60 days")
	}
	u.LastVisitTime = sql.NullTime{Time: time.Now().Add(-61 * 24 * time.Hour), Valid: true}
	if check(u) {
		t.Error("did not expect a block for a visit more than 60 days ago")
	}
}

// TestGetUserBlockers_AggregatesReasonsPerCategory covers the fan-out from
// individual blocker checks into the three category flags/reason lists that
// callers (transfer/visit/promotion eligibility) actually consume.
func TestGetUserBlockers_AggregatesReasonsPerCategory(t *testing.T) {
	// GetUserBlockers itself runs CheckUserBlockers over the full table,
	// which includes DB-backed blockers; exercise the aggregation step
	// (aggregateBlockers) directly instead, with a hand-picked blocker list.
	got := aggregateBlockers([]Blocker{
		{Code: "not_in_division", Message: "not in division", Blocks: []Block{BlocksTransfer, BlocksPromotion}},
		{Code: "visit_low_rating", Message: "low rating", Blocks: []Block{BlocksVisit}},
	})

	if !got.IsTransferBlocked || len(got.TransferBlockedReasons) != 1 {
		t.Errorf("expected exactly one transfer block reason, got %+v", got.TransferBlockedReasons)
	}
	if !got.IsPromotionBlocked || len(got.PromotionBlockedReasons) != 1 {
		t.Errorf("expected exactly one promotion block reason, got %+v", got.PromotionBlockedReasons)
	}
	if !got.IsVisitBlocked || len(got.VisitBlockedReasons) != 1 {
		t.Errorf("expected exactly one visit block reason, got %+v", got.VisitBlockedReasons)
	}
}
