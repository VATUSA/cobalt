package user_migration

import (
	"context"
	"database/sql"
	"log"
	"time"
	"vatusa-cobalt/db"
	"vatusa-cobalt/dbconn"
)

func BulkMigrateUsers() error {
	queries := dbconn.Queries()
	ctx := context.Background()

	lastPromotionMap := map[int]time.Time{}
	lastTransferMap := map[int]time.Time{}
	lastCompetencyMap := map[int]time.Time{}
	controllerRatingMap := map[int]int{}
	instructorTAMap := map[int]int{}

	lastPromotions, err := queries.GetLegacyLastPromotionDatesToMigrate(ctx)
	if err != nil {
		return err
	}
	for _, lp := range lastPromotions {
		lastPromotionMap[int(lp.Cid)] = lp.LastPromotionDate
	}
	lastPromotions = nil

	lastTransfers, err := queries.GetLegacyLastTransferDateToMigrate(ctx)
	if err != nil {
		return err
	}
	for _, lt := range lastTransfers {
		lastTransferMap[int(lt.Cid)] = lt.LastTransferDate
	}
	lastTransfers = nil

	lastCompetencys, err := queries.GetLegacyLastCompetencyDateToMigrate(ctx)
	if err != nil {
		return err
	}
	for _, ct := range lastCompetencys {
		lastCompetencyMap[int(ct.Cid)] = ct.LastCompetencyDate
	}
	lastCompetencys = nil

	controllerRatings, err := queries.GetLegacyControllerRatingFromPromotionsToMigrate(ctx)
	if err != nil {
		return err
	}
	for _, r := range controllerRatings {
		controllerRatingMap[int(r.Cid)] = int(r.MaxRating)
	}
	controllerRatings = nil

	instructorTA, err := queries.GetLegacyInstructorTA(ctx)
	if err != nil {
		return err
	}
	for _, r := range instructorTA {
		if r.Role == "TA" {
			instructorTAMap[int(r.Cid)] = 10
		} else if r.Role == "INS" {
			instructorTAMap[int(r.Cid)] = 8
		}
	}
	instructorTA = nil

	controllersToMigrate, err := queries.GetLegacyControllersToMigrate(ctx)
	if err != nil {
		return err
	}

	for _, c := range controllersToMigrate {
		params := db.UpsertUserForMigrationParams{
			Cid:                int64(c.Cid),
			ControllerRating:   0,
			InstructorRating:   0,
			Facility:           c.Facility,
			DiscordID:          "",
			LastPromotionTime:  sql.NullTime{},
			LastTransferTime:   sql.NullTime{},
			LastCompetencyDate: sql.NullTime{},
		}

		lastPromotion, ok := lastPromotionMap[int(c.Cid)]
		if ok {
			params.LastPromotionTime = sql.NullTime{
				Time:  lastPromotion,
				Valid: true,
			}
		}

		lastTransfer, ok := lastTransferMap[int(c.Cid)]
		if ok {
			params.LastTransferTime = sql.NullTime{
				Time:  lastTransfer,
				Valid: true,
			}
		}

		lastCompetency, ok := lastCompetencyMap[int(c.Cid)]
		if ok {
			params.LastCompetencyDate = sql.NullTime{
				Time:  lastCompetency,
				Valid: true,
			}
		}

		rating := int(c.Rating)
		if rating > 7 {
			lookup, ok := controllerRatingMap[int(c.Cid)]
			if ok {
				rating = lookup
			} else {
				if rating < 11 {
					rating = 5
					if c.Facility != "ZZN" {
						log.Printf("Unable to find rating history for %d -- Defaulting to C1", c.Cid)
					}
				} else {
					rating = 0
					log.Printf("Unable to find rating history for %d -- Defaulting to NULL", c.Cid)
				}
			}
		}

		if rating >= 2 {
			params.ControllerRating = int32(rating)
		}

		if c.Rating == 8 || c.Rating == 10 {
			params.InstructorRating = c.Rating
		} else if c.Rating > 10 {
			instructorLookup, ok := instructorTAMap[int(c.Cid)]
			if ok {
				params.InstructorRating = int32(instructorLookup)
			}
		}

		err = queries.UpsertUserForMigration(ctx, params)
		if err != nil {
			return err
		}
	}
	return nil
}
