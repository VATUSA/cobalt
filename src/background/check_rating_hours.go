package background

import (
	"errors"
	"log"
	"strconv"
	"vatusa-cobalt/dbconn"
	"vatusa-cobalt/vatsim"
)

func CheckRatingHours(args []string) error {
	if len(args) >= 1 {
		cid, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}
		user, err := dbconn.GetCombinedUserByCID(cid)
		if err != nil {
			return err
		}
		if user == nil {
			return errors.New("user not found")
		}
		err = vatsim.CheckRatingHoursForUser(*user)
		if err != nil {
			return err
		}
		return nil
	}

	users, err := dbconn.GetUsersToCheckRatingHours()
	if err != nil {
		return err
	}
	total := len(users)
	count := 1

	for _, user := range users {
		log.Printf("[%d/%d] Checking hours for user %d", count, total, user.Cid)
		err = vatsim.CheckRatingHoursForUser(user)
		if err != nil {
			log.Printf("Error checking rating hours for user %d: %v", user.Cid, err)
		}
		count++
	}

	// TODO
	return nil
}
