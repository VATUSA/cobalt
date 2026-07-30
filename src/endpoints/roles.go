package endpoints

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"time"
	"vatusa-cobalt/acl"
	"vatusa-cobalt/action"
	"vatusa-cobalt/auth"
	"vatusa-cobalt/config"
	"vatusa-cobalt/db"
	"vatusa-cobalt/dbconn"
	"vatusa-cobalt/models"

	"github.com/labstack/echo/v5"
)

func GetACETeam(c *echo.Context) error {
	users, err := dbconn.GetUsersByRole(string(acl.RoleACETeam), nil)
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}
	output := models.ACETeamMembersFromDatabase(users)
	return c.JSON(http.StatusOK, output)
}

func checkRoleManagePerm(c *echo.Context, facility string, object acl.Object) bool {
	if facility == "ZHQ" {
		return AssertGlobal(c, object, acl.ActionWrite)
	}
	return AssertFacility(c, facility, object, acl.ActionWrite)
}

func validateRoleFacility(role acl.Role, facility string) (acl.Object, error) {
	if slices.Contains(acl.AutomaticRoles, role) {
		return "", errors.New("cannot manage automatic roles")
	}

	object, ok := acl.RoleToPermissionObjectMap[role]
	if !ok {
		return "", errors.New("unknown role")
	}

	if config.IsSpecialFacility(facility) {
		return "", errors.New("cannot manage roles for this facility")
	}

	if facility == "ZHQ" && !slices.Contains(acl.GlobalRoles, role) {
		return "", errors.New("role is not a global role")
	}

	if facility != "ZHQ" && !slices.Contains(acl.FacilityRoles, role) {
		return "", errors.New("role is not a facility role")
	}

	return object, nil
}

func GrantRole(c *echo.Context) error {
	var req models.GrantRoleRequest
	if err := c.Bind(&req); err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}

	cidStr := c.Param("cid")
	facility := c.Param("facility")

	cidInt, err := strconv.Atoi(cidStr)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("invalid cid"))
	}

	roleStr := req.Role
	role := acl.Role(roleStr)
	if role == "" {
		return GenericError(c, http.StatusBadRequest, errors.New("invalid role"))
	}

	object, err := validateRoleFacility(role, facility)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}

	_, err = dbconn.Queries().GetFacility(context.Background(), facility)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("unknown facility"))
	}

	if !auth.IsLoggedIn(c) {
		return ErrorNoPermission(c)
	}

	if !checkRoleManagePerm(c, facility, object) {
		return nil
	}

	user, err := dbconn.GetCombinedUserByCID(cidInt)
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}
	if user == nil {
		return GenericError(c, http.StatusNotFound, errors.New("user not found"))
	}

	dbFacility := facility
	if facility == "ZHQ" {
		dbFacility = acl.ScopedRoleGlobalFacility
	}

	existingRoles, err := dbconn.Queries().GetRolesForUser(context.Background(), int32(cidInt))
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}
	for _, r := range existingRoles {
		if r.Facility == dbFacility && r.Role == roleStr {
			return GenericError(c, http.StatusConflict, errors.New("role already assigned"))
		}
	}

	err = dbconn.Queries().AddRoleToUser(context.Background(), db.AddRoleToUserParams{
		Cid:        int32(cidInt),
		Facility:   dbFacility,
		Role:       roleStr,
		GrantorCid: int32(auth.GetUserCid(c)),
		GrantedAt:  time.Now().Unix(),
	})
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}

	action.Log(*user, action.RoleGranted, fmt.Sprintf("Granted role %s at %s", roleStr, facility), int64(auth.GetUserCid(c)))

	return GenericSuccess(c, cidInt)
}

func RevokeRole(c *echo.Context) error {
	cidStr := c.Param("cid")
	facility := c.Param("facility")
	roleStr := c.Param("role")

	cidInt, err := strconv.Atoi(cidStr)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("invalid cid"))
	}

	role := acl.Role(roleStr)
	if role == "" {
		return GenericError(c, http.StatusBadRequest, errors.New("invalid role"))
	}

	object, err := validateRoleFacility(role, facility)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}

	_, err = dbconn.Queries().GetFacility(context.Background(), facility)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("unknown facility"))
	}

	if !auth.IsLoggedIn(c) {
		return ErrorNoPermission(c)
	}

	if !checkRoleManagePerm(c, facility, object) {
		return nil
	}

	user, err := dbconn.GetCombinedUserByCID(cidInt)
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}
	if user == nil {
		return GenericError(c, http.StatusNotFound, errors.New("user not found"))
	}

	dbFacility := facility
	if facility == "ZHQ" {
		dbFacility = acl.ScopedRoleGlobalFacility
	}

	existingRoles, err := dbconn.Queries().GetRolesForUser(context.Background(), int32(cidInt))
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}
	found := false
	for _, r := range existingRoles {
		if r.Facility == dbFacility && r.Role == roleStr {
			found = true
			break
		}
	}
	if !found {
		return GenericError(c, http.StatusNotFound, errors.New("role not found"))
	}

	err = dbconn.Queries().RemoveRoleFromUser(context.Background(), db.RemoveRoleFromUserParams{
		Cid:      int32(cidInt),
		Facility: dbFacility,
		Role:     roleStr,
	})
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}

	action.Log(*user, action.RoleRevoked, fmt.Sprintf("Revoked role %s at %s", roleStr, facility), int64(auth.GetUserCid(c)))

	return GenericSuccess(c, cidInt)
}
