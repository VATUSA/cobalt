package vatsim

import (
	"context"
	"database/sql"
	"log"
	"time"
	"vatusa-cobalt/db"
	"vatusa-cobalt/dbconn"
)

func SyncDivisionMembers() error {
	i := 1
	for {
		log.Printf("Fetching page %d\n", i)
		divisionMembers, err := GetDivisionMembersPage(i)
		if err != nil {
			return err
		}
		if len(divisionMembers.Items) == 0 {
			break
		}

		for _, member := range divisionMembers.Items {
			err = ProcessMemberData(member)
			if err != nil {
				log.Printf("Error processing member %d: %s\n", member.Id, err)
				// Don't return on error here, as we want to finish batch processing
			}
		}
		i++
	}
	return nil
}

func SyncByCID(cid int) error {
	member, err := GetMemberDataByCid(cid)
	if err != nil {
		return err
	}
	return ProcessMemberData(*member)
}

func ProcessMemberData(member MemberData) error {
	queries := dbconn.Queries()
	ctx := context.Background()

	regTime, err := time.Parse("2006-01-02T15:04:05", member.RegistrationDate)
	if err != nil {
		return err
	}

	params := db.UpsertVatsimUserParams{
		Cid:            int64(member.Id),
		NameFirst:      member.NameFirst,
		NameLast:       member.NameLast,
		Email:          member.Email,
		Rating:         int32(member.Rating),
		Pilotrating:    int32(member.PilotRating),
		Militaryrating: int32(member.MilitaryRating),
		SuspendDate:    sql.NullTime{},
		RegistrationDate: sql.NullTime{
			Time:  regTime,
			Valid: true,
		},
		RegionID:           member.RegionId,
		DivisionID:         member.DivisionId,
		SubdivisionID:      sql.NullString{},
		LatestRatingChange: sql.NullTime{},
		LastSync:           time.Now(),
	}
	if member.SubdivisionId != nil {
		params.SubdivisionID = sql.NullString{
			String: *member.SubdivisionId,
			Valid:  true,
		}
	}
	if member.SuspensionDate != nil {
		parse, err := time.Parse("2006-01-02T15:04:05", *member.SuspensionDate)
		if err != nil {
			return err
		}
		params.SuspendDate = sql.NullTime{
			Time:  parse,
			Valid: true,
		}
	}

	if member.LastRatingChange != nil {
		parse, err := time.Parse("2006-01-02T15:04:05", *member.LastRatingChange)
		if err != nil {
			return err
		}
		params.LatestRatingChange = sql.NullTime{
			Time:  parse,
			Valid: true,
		}
	}
	err = queries.UpsertVatsimUser(ctx, params)
	if err != nil {
		return err
	}

	// TODO: Add checks for suspended, transferred in, transferred out, etc
	return nil
}
