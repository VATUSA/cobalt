package auth

import (
	"errors"
	"time"
	"vatusa-cobalt/config"

	"github.com/golang-jwt/jwt/v5"
)

func CreateToken(cid int, displayName string, globalPermissions string, facilityPermissions string, isStaff bool) (string, error) {
	key := config.JWTKey()
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss":                  "cobalt",
		"cid":                  cid,
		"display_name":         displayName,
		"global_permissions":   globalPermissions,
		"facility_permissions": facilityPermissions,
		"is_staff":             isStaff,
		"exp":                  jwt.NewNumericDate(time.Now().Add(SessionDuration)),
	})
	s, err := t.SignedString(key)
	return s, err
}

func GetCIDFromToken(token string) (int, error) {
	key := config.JWTKey()
	t, err := jwt.Parse(
		token,
		func(token *jwt.Token) (any, error) {
			return key, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return -1, err
	}
	claims, ok := t.Claims.(jwt.MapClaims)
	if ok && t.Valid {
		return int(claims["cid"].(float64)), nil
	} else {
		return -1, errors.New("invalid token")
	}
}
