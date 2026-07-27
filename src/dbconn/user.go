package dbconn

import (
	"context"
	"database/sql"
	"errors"
	"time"
	"vatusa-cobalt/db"
)

func GetCombinedUserByCID(cid int) (*db.GetCombinedUserRow, error) {
	ctx := context.Background()
	params := db.GetCombinedUserParams{
		Cid: sql.NullInt64{
			Int64: int64(cid),
			Valid: true,
		},
	}
	users, err := Queries().GetCombinedUser(ctx, params)
	if err != nil {
		return nil, err
	}
	if len(users) > 0 {
		return &users[0], nil
	}
	return nil, nil
}

func GetCombinedUsersByHomeFacility(facility string) ([]db.GetCombinedUserRow, error) {
	ctx := context.Background()
	params := db.GetCombinedUserParams{
		HomeFacility: sql.NullString{
			String: facility,
			Valid:  true,
		},
	}
	users, err := Queries().GetCombinedUser(ctx, params)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func GetCombinedUsersByVisitingFacility(facility string) ([]db.GetCombinedUserRow, error) {
	ctx := context.Background()
	params := db.GetCombinedUserParams{
		VisitFacility: sql.NullString{
			String: facility,
			Valid:  true,
		},
	}
	users, err := Queries().GetCombinedUser(ctx, params)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func GetUserRatingHours(cid int, rating int) (*db.UserRatingHour, error) {
	ctx := context.Background()
	params := db.GetUserRatingHoursParams{
		Cid:    int64(cid),
		Rating: int32(rating),
	}
	result, err := Queries().GetUserRatingHours(ctx, params)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func StoreUserRatingHours(cid int, rating int, hours int) error {
	ctx := context.Background()
	params := db.StoreUserRatingHoursParams{
		Cid:           int64(cid),
		Rating:        int32(rating),
		Hours:         int32(hours),
		LastCheckTime: time.Now(),
	}
	err := Queries().StoreUserRatingHours(ctx, params)
	return err
}

func GetUsersByRole(role string, facility *string) ([]db.GetUsersByRoleRow, error) {
	ctx := context.Background()
	var fac sql.NullString
	if facility != nil {
		fac = sql.NullString{String: *facility, Valid: true}
	}
	return Queries().GetUsersByRole(ctx, db.GetUsersByRoleParams{
		Role:     role,
		Facility: fac,
	})
}

func GetUsersToCheckRatingHours() ([]db.GetCombinedUserRow, error) {
	ctx := context.Background()
	cids, err := Queries().GetUserRatingHourCheckCids(ctx, 50)
	if err != nil {
		return nil, err
	}

	params := db.GetCombinedUserParams{
		Cids:         cids,
		HasCidsSlice: true,
	}
	users, err := Queries().GetCombinedUser(ctx, params)
	if err != nil {
		return nil, err
	}
	return users, nil
}
