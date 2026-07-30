package config

const (
	FacilityAcademy   = "ZAE"
	FacilityNonMember = "ZZN"
	FacilityInactive  = "ZZI"
	FacilityNotExists = "ZZZ"
)

const (
	RegionAmericas = "AMAS"
	DivisionVATUSA = "USA"
)

func IsSpecialFacility(facility string) bool {
	switch facility {
	case FacilityAcademy, FacilityNonMember, FacilityInactive, FacilityNotExists:
		return true
	}
	return false
}
