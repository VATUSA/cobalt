package vatsim

import (
	"log"
	"vatusa-cobalt/config"
	"vatusa-cobalt/db"
	"vatusa-cobalt/moodle"
)

// ratingCohorts maps a VATSIM rating id to the idnumber of its corresponding
// Moodle rating cohort. Ratings with no entry (AFK/SUS) are not synced.
var ratingCohorts = map[int32]string{
	config.RatingObserver:         "OBS",
	config.RatingStudent1:         "S1",
	config.RatingStudent2:         "S2",
	config.RatingStudent3:         "S3",
	config.RatingController1:      "C1",
	6:                             "C2",
	config.RatingController3:      "C3",
	config.RatingInstructor:       "I1",
	9:                             "I2",
	config.RatingSeniorInstructor: "I3",
	config.RatingSupervisor:       "SUP",
	config.RatingAdministrator:    "ADM",
}

// syncMoodleCohort keeps a user's Moodle rating cohort (e.g. "OBS" for a new
// Observer) current on every login. This is the replacement for the legacy
// api repo's ULSHelper::doHandleLogin, which queued this work on every login
// through api's own SSO routes -- a path nothing calls anymore now that
// cobalt/webapps is the primary login flow. The scheduled api moodle:sync
// job remains as a backstop for changes that don't happen at login time
// (facility/staff cohort changes), but new controllers must not depend on it
// to get Observer access.
//
// Errors are logged, not returned: Moodle being slow or unreachable must
// never fail a login.
func syncMoodleCohort(vatsimUser db.VatsimUser) {
	if !config.MoodleEnabled() {
		return
	}
	cohort, ok := ratingCohorts[vatsimUser.Rating]
	if !ok {
		return
	}

	moodleID, exists, err := moodle.GetUserID(vatsimUser.Cid)
	if err != nil {
		log.Printf("moodle: error looking up user for cid %d: %v", vatsimUser.Cid, err)
		return
	}
	if !exists {
		moodleID, err = moodle.CreateUser(vatsimUser.Cid, vatsimUser.NameFirst, vatsimUser.NameLast, vatsimUser.Email)
		if err != nil {
			log.Printf("moodle: error creating user for cid %d: %v", vatsimUser.Cid, err)
			return
		}
	}

	if err := moodle.AssignCohort(moodleID, cohort); err != nil {
		log.Printf("moodle: error assigning cohort %q to cid %d: %v", cohort, vatsimUser.Cid, err)
	}
}
