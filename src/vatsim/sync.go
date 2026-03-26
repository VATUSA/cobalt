package vatsim

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"time"
	"vatusa-cobalt/config"
	"vatusa-cobalt/db"
	"vatusa-cobalt/dbconn"
	"vatusa-cobalt/roster"
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

	regTime, err := time.Parse(config.ConnectTimestampFormat, member.RegistrationDate)
	if err != nil {
		return err
	}

	var user *db.GetUserByCIDRow
	user_, err := queries.GetUserByCID(ctx, int64(member.Id))
	if err == nil {
		user = &user_
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
		parse, err := time.Parse(config.ConnectTimestampFormat, *member.SuspensionDate)
		if err != nil {
			return err
		}
		params.SuspendDate = sql.NullTime{
			Time:  parse,
			Valid: true,
		}
	}

	if member.LastRatingChange != nil {
		parse, err := time.Parse(config.ConnectTimestampFormat, *member.LastRatingChange)
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

	if user != nil {
		if member.Rating == config.RatingInactive {
			// Facility = config.FacilityInactive
			if user.Facility != config.FacilityInactive {
				err = setFacility(user, config.FacilityInactive)
				if err != nil {
					return err
				}
				err = removeVisits(user)
				if err != nil {
					return err
				}
			}
		} else if member.Rating == config.RatingSuspended {
			// Facility = config.FacilityInactive
			if user.Facility != config.FacilityInactive {
				err = setFacility(user, config.FacilityInactive)
				if err != nil {
					return err
				}
				err = removeVisits(user)
				if err != nil {
					return err
				}
				// TODO: Send suspended email?
			}
		} else if member.RegionId == "AMAS" && member.DivisionId == "USA" {
			// In division, set facility = config.FacilityAcademy if necessary
			if user.Facility == config.FacilityInactive || user.Facility == config.FacilityNonMember {
				err = setFacility(user, config.FacilityAcademy)
				if err != nil {
					return err
				}
			}
		} else {
			// Facility = config.FacilityNonMember
			if user.Facility != config.FacilityNonMember {
				err = setFacility(user, config.FacilityNonMember)
				if err != nil {
					return err
				}
			}
		}
	}

	// TODO: Add checks for suspended, transferred in, transferred out, etc
	return nil
}

func setFacility(user *db.GetUserByCIDRow, facility string) error {
	return roster.ForceTransfer(user.Cid, user.Facility, facility, "VATSIM Sync", 0)
}

func removeVisits(user *db.GetUserByCIDRow) error {
	if user.VisitingFacilities.Valid {
		visitingFacilities := strings.Split(user.VisitingFacilities.String, ",")
		for _, _ = range visitingFacilities {
			// TODO: Remove user from visiting facility
		}
	}
	return nil
}
