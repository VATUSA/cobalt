package vatsim

import (
	"encoding/json"
	"fmt"
	"vatusa-cobalt/config"
	"vatusa-cobalt/db"
	"vatusa-cobalt/dbconn"
)

func GetMemberStatisticsByCid(cid int) (*MemberStatistics, error) {
	uri := fmt.Sprintf("/members/%d/stats", cid)
	data, err := API2Request(uri, "GET", nil)
	if err != nil {
		return nil, err
	}

	var stats MemberStatistics
	err = json.Unmarshal(data, &stats)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

func CheckRatingHoursForUser(user db.GetCombinedUserRow) error {
	if user.ControllerRating < config.RatingStudent1 {
		return nil
	}
	stats, err := GetMemberStatisticsByCid(int(user.Cid))
	if err != nil {
		return err
	}
	rating := user.ControllerRating
	if rating > config.RatingController1 {
		rating = config.RatingController1
	}
	hours := getHoursTotalForRating(stats, int(rating))
	err = dbconn.StoreUserRatingHours(int(user.Cid), int(rating), int(hours))
	return err
}

func getHoursTotalForRating(stats *MemberStatistics, rating int) float64 {
	switch rating {
	case config.RatingStudent1:
		return stats.S1
	case config.RatingStudent2:
		return stats.S2
	case config.RatingStudent3:
		return stats.S3
	case config.RatingController1:
		return stats.C1 + stats.C2 + stats.C3 + stats.I1 + stats.I2 + stats.I3
	default:
		return 0
	}
}
