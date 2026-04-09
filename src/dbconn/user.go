package dbconn

import (
	"context"
	"database/sql"
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
