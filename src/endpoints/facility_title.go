package endpoints

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"vatusa-cobalt/acl"
	"vatusa-cobalt/action"
	"vatusa-cobalt/auth"
	"vatusa-cobalt/db"
	"vatusa-cobalt/dbconn"
	"vatusa-cobalt/models"

	"github.com/labstack/echo/v5"
)

func GetFacilityTitles(c *echo.Context) error {
	facility := c.Param("facility")

	titles, err := dbconn.Queries().GetFacilityTitles(c.Request().Context(), facility)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, models.FacilityTitlesFromDatabase(titles))
}

func validateTitleFacility(c *echo.Context, facility string) bool {
	if _, err := dbconn.Queries().GetFacility(c.Request().Context(), facility); err != nil {
		_ = RespondError(c, http.StatusBadRequest, errors.New("unknown facility"))
		return false
	}
	return true
}

func CreateFacilityTitle(c *echo.Context) error {
	facility := c.Param("facility")
	if !AssertGlobal(c, acl.ObjectFacilityTitleManagement, acl.ActionWrite) {
		return nil
	}
	if !validateTitleFacility(c, facility) {
		return nil
	}
	ctx := c.Request().Context()
	var request models.FacilityTitleRequest
	err := c.Bind(&request)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, err)
	}
	if _, ok := acl.TitleTierToPermissionObjectMap[request.Tier]; !ok {
		return RespondError(c, http.StatusBadRequest, errors.New("invalid title tier"))
	}

	result, err := dbconn.Queries().CreateFacilityTitle(ctx, db.CreateFacilityTitleParams{
		Facility:  facility,
		Title:     request.Title,
		Code:      request.Code,
		Tier:      request.Tier,
		CreatedAt: time.Now().Unix(),
	})
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, err)
	}
	lastInsertId, err := result.LastInsertId()
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, err)
	}

	return RespondSuccess(c, int(lastInsertId))
}

func DeleteFacilityTitle(c *echo.Context) error {
	facility := c.Param("facility")
	if !AssertGlobal(c, acl.ObjectFacilityTitleManagement, acl.ActionWrite) {
		return nil
	}
	if !validateTitleFacility(c, facility) {
		return nil
	}
	ctx := c.Request().Context()
	titleId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, errors.New("invalid title id"))
	}

	title, err := dbconn.Queries().GetFacilityTitleById(ctx, titleId)
	if err != nil {
		return RespondError(c, http.StatusNotFound, errors.New("title not found"))
	}
	if title.Facility != facility {
		return RespondError(c, http.StatusNotFound, errors.New("title not found"))
	}

	err = dbconn.Queries().DeleteFacilityTitleById(ctx, titleId)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, SafeError("failed to delete title"), err)
	}
	return RespondSuccess(c, int(titleId))
}

func GetUserFacilityTitles(c *echo.Context) error {
	facility := c.Param("facility")
	cid, err := strconv.ParseInt(c.Param("cid"), 10, 32)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, errors.New("invalid cid"))
	}

	titles, err := dbconn.Queries().GetUserTitlesByFacility(c.Request().Context(), db.GetUserTitlesByFacilityParams{
		Cid:      int32(cid),
		Facility: facility,
	})
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, models.UserFacilityTitlesFromDatabase(titles))
}

func AssignUserFacilityTitles(c *echo.Context) error {
	facility := c.Param("facility")
	if !validateTitleFacility(c, facility) {
		return nil
	}
	ctx := c.Request().Context()
	cid, err := strconv.ParseInt(c.Param("cid"), 10, 32)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, errors.New("invalid cid"))
	}
	var request models.AssignFacilityTitleRequest
	err = c.Bind(&request)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, err)
	}
	titleId := request.TitleId

	title, err := dbconn.Queries().GetFacilityTitleById(ctx, titleId)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, errors.New("invalid title id"))
	}
	if title.Facility != facility {
		return RespondError(c, http.StatusBadRequest, errors.New("title does not belong to facility"))
	}
	permissionObject, ok := acl.TitleTierToPermissionObjectMap[title.Tier]
	if !ok {
		return RespondError(c, http.StatusBadRequest, errors.New("title tier cannot be assigned"))
	}
	if !AssertFacility(c, facility, permissionObject, acl.ActionWrite) {
		return nil
	}

	user, err := dbconn.GetCombinedUserByCID(int(cid))
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, err)
	}
	if user == nil {
		return RespondError(c, http.StatusNotFound, errors.New("user not found"))
	}

	grantorCid := int32(auth.GetUserCid(c))
	grantedAt := time.Now().Unix()
	err = dbconn.Queries().AssignUserTitle(ctx, db.AssignUserTitleParams{
		Cid:        int32(cid),
		TitleID:    titleId,
		GrantorCid: grantorCid,
		GrantedAt:  grantedAt,
	})
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, err)
	}

	action.Log(dbconn.Queries(), *user, action.TitleGranted, fmt.Sprintf("Granted title %s at %s", title.Title, facility), int64(auth.GetUserCid(c)))

	return RespondSuccess(c, int(cid))
}

func DeleteUserFacilityTitle(c *echo.Context) error {
	facility := c.Param("facility")
	if !validateTitleFacility(c, facility) {
		return nil
	}
	ctx := c.Request().Context()
	cid, err := strconv.ParseInt(c.Param("cid"), 10, 32)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, errors.New("invalid cid"))
	}
	titleId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, errors.New("invalid title id"))
	}

	title, err := dbconn.Queries().GetFacilityTitleById(ctx, titleId)
	if err != nil {
		return RespondError(c, http.StatusNotFound, errors.New("title not found"))
	}
	if title.Facility != facility {
		return RespondError(c, http.StatusNotFound, errors.New("title not found"))
	}
	permissionObject, ok := acl.TitleTierToPermissionObjectMap[title.Tier]
	if !ok {
		return RespondError(c, http.StatusBadRequest, errors.New("title tier cannot be assigned"))
	}
	if !AssertFacility(c, facility, permissionObject, acl.ActionWrite) {
		return nil
	}

	user, err := dbconn.GetCombinedUserByCID(int(cid))
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, err)
	}
	if user == nil {
		return RespondError(c, http.StatusNotFound, errors.New("user not found"))
	}

	err = dbconn.Queries().DeleteUserTitle(ctx, db.DeleteUserTitleParams{
		Cid:     int32(cid),
		TitleID: titleId,
	})
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, SafeError("failed to remove title"), err)
	}

	action.Log(dbconn.Queries(), *user, action.TitleRevoked, fmt.Sprintf("Revoked title %s at %s", title.Title, facility), int64(auth.GetUserCid(c)))

	return RespondSuccess(c, int(titleId))
}
