package vatsim

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"vatusa-cobalt/config"
	"vatusa-cobalt/db"
	"vatusa-cobalt/dbconn"
	"vatusa-cobalt/roster"
)

func VatsimUserUpdated(cid int64) error {
	vatsimUser, err := dbconn.Queries().GetVatsimUserByCid(context.Background(), cid)
	if err != nil {
		return err
	}
	displayName := vatsimDisplayName(vatsimUser)
	user, err := dbconn.Queries().GetUserByCID(context.Background(), vatsimUser.Cid)
	if errors.Is(err, sql.ErrNoRows) {
		if vatsimUser.Rating == config.RatingInactive || vatsimUser.Rating == config.RatingSuspended {
			return nil
		}
		params := db.InsertUserFromVatsimSyncParams{
			Cid:         cid,
			DisplayName: displayName,
			Facility:    "",
		}
		if vatsimUser.RegionID == "AMAS" && vatsimUser.DivisionID == "USA" {
			params.Facility = config.FacilityAcademy
		} else {
			params.Facility = config.FacilityNonMember
		}
		err = dbconn.Queries().InsertUserFromVatsimSync(context.Background(), params)
		if err != nil {
			return err
		}
		user, err = dbconn.Queries().GetUserByCID(context.Background(), vatsimUser.Cid)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else if user.DisplayName != displayName {
		// Existing user: keep display_name current with the VATSIM record. Prior to
		// this the existing-user branch returned early, so a migrated user's name was
		// never populated and facility/suspension was never re-evaluated on sync.
		err = dbconn.Queries().UpdateUserFromVatsimSync(context.Background(), db.UpdateUserFromVatsimSyncParams{
			Cid:         vatsimUser.Cid,
			DisplayName: displayName,
		})
		if err != nil {
			return err
		}
		user.DisplayName = displayName
	}
	return onVatsimUserUpdate(vatsimUser, user)
}

func vatsimDisplayName(vatsimUser db.VatsimUser) string {
	return strings.TrimSpace(vatsimUser.NameFirst + " " + vatsimUser.NameLast)
}

func onVatsimUserUpdate(vatsimUser db.VatsimUser, user db.GetUserByCIDRow) error {
	err := handleSuspendedAndInactive(vatsimUser, user)
	if err != nil {
		return err
	}
	err = handleUserInDivision(vatsimUser, user)
	if err != nil {
		return err
	}

	return nil
}

func handleUserInDivision(vatsimUser db.VatsimUser, user db.GetUserByCIDRow) error {
	if vatsimUser.RegionID == "AMAS" && vatsimUser.DivisionID == "USA" {
		if user.Facility == config.FacilityInactive || user.Facility == config.FacilityNonMember {
			err := setFacility(user, config.FacilityAcademy)
			if err != nil {
				return err
			}
			// TODO: Send welcome email
		}
	} else {
		if user.Facility != config.FacilityNonMember {
			err := setFacility(user, config.FacilityNonMember)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func handleSuspendedAndInactive(vatsimUser db.VatsimUser, user db.GetUserByCIDRow) error {
	if vatsimUser.Rating == config.RatingInactive || vatsimUser.Rating == config.RatingSuspended {
		if user.Facility != config.FacilityInactive {
			if vatsimUser.Rating == config.RatingSuspended {
				// TODO: Send suspended email
			}
			err := setFacility(user, config.FacilityInactive)
			if err != nil {
				return err
			}
			err = removeVisits(user)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func setFacility(user db.GetUserByCIDRow, facility string) error {
	combinedUser, err := dbconn.GetCombinedUserByCID(int(user.Cid))
	if err != nil {
		return err
	}
	if combinedUser == nil {
		return errors.New("user not found")
	}
	return roster.ForceTransfer(*combinedUser, user.Facility, facility, "VATSIM Sync", 0)
}

func removeVisits(user db.GetUserByCIDRow) error {
	if user.VisitingFacilities.Valid {
		visitingFacilities := strings.Split(user.VisitingFacilities.String, ",")
		for _, _ = range visitingFacilities {
			// TODO: Remove user from visiting facility
		}
	}
	return nil
}
