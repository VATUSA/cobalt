package dbconn

import (
	"cmp"
	"context"
	"database/sql"
	"errors"
	"slices"
	"strings"
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

// SearchUsers resolves a free-text query against the user base. An all-digit
// query is treated as a possibly-partial cid; a single word matches against
// either name; two or more words match a first and last name together. All
// matching is prefix-only so it can use the vatsim_user name indexes.
//
// Matches are resolved to cids first and hydrated through GetCombinedUser, so
// search results are the same shape as every other user response.
func SearchUsers(query string, limit int) ([]db.GetCombinedUserRow, error) {
	params, ok := buildUserSearchParams(query, limit)
	if !ok {
		return nil, nil
	}
	ctx := context.Background()
	cids, err := Queries().SearchUserCids(ctx, params)
	if err != nil {
		return nil, err
	}
	if len(cids) == 0 {
		return nil, nil
	}
	users, err := Queries().GetCombinedUser(ctx, db.GetCombinedUserParams{
		Cids:         cids,
		HasCidsSlice: true,
	})
	if err != nil {
		return nil, err
	}
	// GetCombinedUser is unordered, so restore the order SearchUserCids applied
	// the limit in. Compared case-insensitively to approximate the column collation.
	slices.SortFunc(users, func(a, b db.GetCombinedUserRow) int {
		return cmp.Or(
			strings.Compare(strings.ToLower(a.NameLast), strings.ToLower(b.NameLast)),
			strings.Compare(strings.ToLower(a.NameFirst), strings.ToLower(b.NameFirst)),
			cmp.Compare(a.Cid, b.Cid),
		)
	})
	return users, nil
}

// likeEscaper neutralizes the LIKE wildcards in caller-supplied input, so a
// query containing % or _ stays a literal prefix search instead of turning
// into a full scan.
var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

func likePrefix(s string) sql.NullString {
	return sql.NullString{String: likeEscaper.Replace(s) + "%", Valid: true}
}

// buildUserSearchParams reports false when the query holds nothing searchable,
// which callers treat as an empty result rather than an unfiltered one.
func buildUserSearchParams(query string, limit int) (db.SearchUserCidsParams, bool) {
	params := db.SearchUserCidsParams{Limit: int32(limit)}
	fields := strings.Fields(query)
	if len(fields) == 0 {
		return params, false
	}
	if len(fields) == 1 {
		if isAllDigits(fields[0]) {
			// The query appends the wildcard itself, and digits can't contain
			// one, so this needs no escaping.
			params.CidPrefix = fields[0]
		} else {
			params.NameAny = likePrefix(fields[0])
		}
		return params, true
	}
	params.NameFirst = likePrefix(fields[0])
	params.NameLast = likePrefix(strings.Join(fields[1:], " "))
	return params, true
}

func isAllDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
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
