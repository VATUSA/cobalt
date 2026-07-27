package config

const (
	RatingInactive         = -1
	RatingSuspended        = 0
	RatingObserver         = 1
	RatingStudent1         = 2
	RatingStudent2         = 3
	RatingStudent3         = 4
	RatingController1      = 5
	RatingController3      = 7
	RatingInstructor       = 8
	RatingSeniorInstructor = 10
	RatingSupervisor       = 11
	RatingAdministrator    = 12
)

func RatingShort(rating int) string {
	switch rating {
	case RatingInactive:
		return "INAC"
	case RatingSuspended:
		return "SUS"
	case RatingObserver:
		return "OBS"
	case RatingStudent1:
		return "S1"
	case RatingStudent2:
		return "S2"
	case RatingStudent3:
		return "S3"
	case RatingController1:
		return "C1"
	case RatingController3:
		return "C3"
	case RatingInstructor:
		return "I1"
	case RatingSeniorInstructor:
		return "I3"
	case RatingSupervisor:
		return "SUP"
	case RatingAdministrator:
		return "ADM"
	default:
		return ""
	}
}

func RatingLong(rating int) string {
	switch rating {
	case RatingInactive:
		return "Inactive"
	case RatingSuspended:
		return "Suspended"
	case RatingObserver:
		return "Observer"
	case RatingStudent1:
		return "Tower Trainee"
	case RatingStudent2:
		return "Tower Controller"
	case RatingStudent3:
		return "Senior Student"
	case RatingController1:
		return "Enroute Controller"
	case RatingController3:
		return "Senior Controller"
	case RatingInstructor:
		return "Instructor"
	case RatingSeniorInstructor:
		return "Senior Instructor"
	case RatingSupervisor:
		return "Supervisor"
	case RatingAdministrator:
		return "Administrator"
	default:
		return ""
	}
}
