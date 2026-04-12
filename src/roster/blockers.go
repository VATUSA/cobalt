package roster

import (
	"slices"
	"time"
	"vatusa-cobalt/config"
	"vatusa-cobalt/db"
)

type Block string

const BlocksTransfer Block = "transfer"
const BlocksVisit Block = "visit"
const BlocksPromotion Block = "promotion"

// BlockerCheckFunc should return true if the blocker should be applied
type BlockerCheckFunc func(user db.GetCombinedUserRow) bool

type Blocker struct {
	Code    string
	Message string
	Blocks  []Block
	Check   BlockerCheckFunc
}

type UserBlockers struct {
	IsTransferBlocked       bool     `json:"is_transfer_blocked"`
	TransferBlockedReasons  []string `json:"transfer_blocked_reasons"`
	IsVisitBlocked          bool     `json:"is_visit_blocked"`
	VisitBlockedReasons     []string `json:"visit_blocked_reasons"`
	IsPromotionBlocked      bool     `json:"is_promotion_blocked"`
	PromotionBlockedReasons []string `json:"promotion_blocked_reasons"`
}

var Blockers = []Blocker{
	{
		Code:    "not_in_division",
		Message: "Not an active member of the VATUSA Division",
		Blocks: []Block{
			BlocksTransfer,
			BlocksPromotion,
		},
		Check: func(user db.GetCombinedUserRow) bool {
			return !(user.RegionID == config.RegionAmericas && user.DivisionID == config.DivisionVATUSA)
		},
	},
	{
		Code:    "needs_exam",
		Message: "Needs to complete Basic course or RCE Exam",
		Blocks: []Block{
			BlocksTransfer,
			BlocksVisit,
		},
		Check: func(user db.GetCombinedUserRow) bool {
			if !user.LastCompetencyDate.Valid {
				return true
			}
			return time.Since(user.LastCompetencyDate.Time) > 180*24*time.Hour
		},
	},
	{
		Code:    "recent_transfer",
		Message: "Has transferred in the last 90 days",
		Blocks: []Block{
			BlocksTransfer,
		},
		Check: func(user db.GetCombinedUserRow) bool {
			if !user.LastTransferTime.Valid {
				return false
			}
			return time.Since(user.LastTransferTime.Time) < 90*24*time.Hour
		},
	},
	{
		Code:    "recent_promotion",
		Message: "Has been promoted in the last 90 days",
		Blocks: []Block{
			BlocksTransfer,
			BlocksVisit,
		},
		Check: func(user db.GetCombinedUserRow) bool {
			if !user.LastPromotionTime.Valid {
				return false
			}
			return time.Since(user.LastPromotionTime.Time) < 90*24*time.Hour
		},
	},
	{
		Code:    "missing_rating_hours",
		Message: "Has not completed 50 hours since promotion",
		Blocks: []Block{
			BlocksTransfer,
			BlocksVisit,
		},
		Check: func(user db.GetCombinedUserRow) bool {
			// TODO
			return false
		},
	},
	{
		Code:    "is_staff",
		Message: "Is an active staff member",
		Blocks: []Block{
			BlocksTransfer,
		},
		Check: func(user db.GetCombinedUserRow) bool {
			// TODO
			return false
		},
	},
	{
		Code:    "instructor_rating",
		Message: "Holds an Instructor I1 or I3 rating",
		Blocks: []Block{
			BlocksTransfer,
		},
		Check: func(user db.GetCombinedUserRow) bool {
			return user.Rating == config.RatingInstructor || user.Rating == config.RatingSeniorInstructor || user.InstructorRating > 0
		},
	},
	{
		Code:    "pending_transfer",
		Message: "Has a pending transfer request",
		Blocks: []Block{
			BlocksTransfer,
		},
		Check: func(user db.GetCombinedUserRow) bool {
			// TODO
			return false
		},
	},
	{
		Code:    "in_training_academy",
		Message: "Does not have a home facility (in academy)",
		Blocks: []Block{
			BlocksVisit,
			BlocksPromotion,
		},
		Check: func(user db.GetCombinedUserRow) bool {
			return user.Facility == config.FacilityAcademy
		},
	},
	{
		Code:    "is_inactive",
		Message: "VATSIM account is not active",
		Blocks: []Block{
			BlocksTransfer,
			BlocksVisit,
			BlocksPromotion,
		},
		Check: func(user db.GetCombinedUserRow) bool {
			return user.Rating == config.RatingInactive || user.Rating == config.RatingSuspended
		},
	},
	{
		Code:    "visit_low_rating",
		Message: "Does not have at least an S3 rating",
		Blocks: []Block{
			BlocksVisit,
		},
		Check: func(user db.GetCombinedUserRow) bool {
			return user.Rating < config.RatingStudent3
		},
	},
	{
		Code:    "recent_visit",
		Message: "Has visited a new facility within the last 60 days",
		Blocks: []Block{
			BlocksVisit,
		},
		Check: func(user db.GetCombinedUserRow) bool {
			//if !user.LastVisitTime.Valid {
			//	return false
			//}
			//return time.Since(user.LastVisitTime.Time) < 60*24*time.Hour
			// TODO: Add LastVisitTime field and use that
			return false
		},
	},
}

func GetUserBlockers(user db.GetCombinedUserRow) UserBlockers {
	userBlockers := UserBlockers{
		IsTransferBlocked:       false,
		TransferBlockedReasons:  []string{},
		IsVisitBlocked:          false,
		IsPromotionBlocked:      false,
		PromotionBlockedReasons: []string{},
	}
	blockers := CheckUserBlockers(user)
	for _, blocker := range blockers {
		if slices.Contains(blocker.Blocks, BlocksTransfer) {
			userBlockers.IsTransferBlocked = true
			userBlockers.TransferBlockedReasons = append(userBlockers.TransferBlockedReasons, blocker.Message)
		}
		if slices.Contains(blocker.Blocks, BlocksVisit) {
			userBlockers.IsVisitBlocked = true
			userBlockers.VisitBlockedReasons = append(userBlockers.VisitBlockedReasons, blocker.Message)
		}
		if slices.Contains(blocker.Blocks, BlocksPromotion) {
			userBlockers.IsPromotionBlocked = true
			userBlockers.PromotionBlockedReasons = append(userBlockers.PromotionBlockedReasons, blocker.Message)
		}
	}
	return userBlockers
}

func CheckUserBlockers(user db.GetCombinedUserRow) []Blocker {
	var out []Blocker
	for _, blocker := range Blockers {
		if blocker.Check(user) {
			out = append(out, blocker)
		}
	}
	return out
}
