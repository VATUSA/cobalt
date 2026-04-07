package vatsim

import (
	"context"
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
	user, err := dbconn.Queries().GetUserByCID(context.Background(), vatsimUser.Cid)
	if err != nil {
		// TODO: Create user?
		return nil
	}
	return onVatsimUserUpdate(vatsimUser, user)
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
	return roster.ForceTransfer(user.Cid, user.Facility, facility, "VATSIM Sync", 0)
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
