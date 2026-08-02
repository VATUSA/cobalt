package endpoints

import (
	"errors"
	"net/http"
	"strconv"
	"time"
	"vatusa-cobalt/acl"
	"vatusa-cobalt/auth"
	"vatusa-cobalt/db"
	"vatusa-cobalt/dbconn"
	"vatusa-cobalt/models"

	"github.com/labstack/echo/v5"
)

func GetFacilityTitles(c *echo.Context) error {
	facility := c.Param("facility")
	if !AssertFacility(c, facility, acl.ObjectFacilityTitle, acl.ActionRead) {
		return nil
	}

	titles, err := dbconn.Queries().GetFacilityTitles(c.Request().Context(), facility)
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, models.FacilityTitlesFromDatabase(titles))
}

func validateTitleFacility(c *echo.Context, facility string) bool {
	if facility == "ZHQ" {
		return true
	}
	if _, err := dbconn.Queries().GetFacility(c.Request().Context(), facility); err != nil {
		_ = GenericError(c, http.StatusBadRequest, errors.New("unknown facility"))
		return false
	}
	return true
}

func CreateFacilityTitle(c *echo.Context) error {
	facility := c.Param("facility")
	if !AssertFacility(c, facility, acl.ObjectFacilityTitle, acl.ActionWrite) {
		return nil
	}
	if !validateTitleFacility(c, facility) {
		return nil
	}
	ctx := c.Request().Context()
	var request models.FacilityTitleRequest
	err := c.Bind(&request)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}

	result, err := dbconn.Queries().CreateFacilityTitle(ctx, db.CreateFacilityTitleParams{
		Facility:  facility,
		Title:     request.Title,
		Code:      request.Code,
		CreatedAt: time.Now().Unix(),
	})
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}
	lastInsertId, err := result.LastInsertId()
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}

	return GenericSuccess(c, int(lastInsertId))
}

func DeleteFacilityTitle(c *echo.Context) error {
	facility := c.Param("facility")
	if !AssertFacility(c, facility, acl.ObjectFacilityTitle, acl.ActionWrite) {
		return nil
	}
	if !validateTitleFacility(c, facility) {
		return nil
	}
	ctx := c.Request().Context()
	titleId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("invalid title id"))
	}

	title, err := dbconn.Queries().GetFacilityTitleById(ctx, titleId)
	if err != nil {
		return GenericError(c, http.StatusNotFound, errors.New("title not found"))
	}
	if title.Facility != facility {
		return GenericError(c, http.StatusNotFound, errors.New("title not found"))
	}

	assignedCount, err := dbconn.Queries().CountUserTitlesByTitleId(ctx, titleId)
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}
	if assignedCount > 0 {
		return GenericError(c, http.StatusBadRequest, errors.New("title is assigned to users"))
	}

	err = dbconn.Queries().DeleteFacilityTitleById(ctx, titleId)
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, errors.New("failed to delete title"))
	}
	return GenericSuccess(c, int(titleId))
}

func GetUserFacilityTitles(c *echo.Context) error {
	facility := c.Param("facility")
	if !AssertFacility(c, facility, acl.ObjectFacilityTitle, acl.ActionRead) {
		return nil
	}
	cid, err := strconv.ParseInt(c.Param("cid"), 10, 32)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("invalid cid"))
	}

	titles, err := dbconn.Queries().GetUserTitlesByFacility(c.Request().Context(), db.GetUserTitlesByFacilityParams{
		Cid:      int32(cid),
		Facility: facility,
	})
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, models.UserFacilityTitlesFromDatabase(titles))
}

func AssignUserFacilityTitles(c *echo.Context) error {
	facility := c.Param("facility")
	if !AssertFacility(c, facility, acl.ObjectFacilityTitle, acl.ActionWrite) {
		return nil
	}
	if !validateTitleFacility(c, facility) {
		return nil
	}
	ctx := c.Request().Context()
	cid, err := strconv.ParseInt(c.Param("cid"), 10, 32)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("invalid cid"))
	}
	var request models.AssignTitlesRequest
	err = c.Bind(&request)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}

	for _, titleId := range request.TitleIds {
		title, err := dbconn.Queries().GetFacilityTitleById(ctx, titleId)
		if err != nil {
			return GenericError(c, http.StatusBadRequest, errors.New("invalid title id"))
		}
		if title.Facility != facility {
			return GenericError(c, http.StatusBadRequest, errors.New("title does not belong to facility"))
		}
	}

	grantorCid := int32(auth.GetUserCid(c))
	grantedAt := time.Now().Unix()
	for _, titleId := range request.TitleIds {
		err := dbconn.Queries().AssignUserTitle(ctx, db.AssignUserTitleParams{
			Cid:        int32(cid),
			TitleID:    titleId,
			GrantorCid: grantorCid,
			GrantedAt:  grantedAt,
		})
		if err != nil {
			return GenericError(c, http.StatusInternalServerError, err)
		}
	}

	return GenericSuccess(c, int(cid))
}

func DeleteUserFacilityTitle(c *echo.Context) error {
	facility := c.Param("facility")
	if !AssertFacility(c, facility, acl.ObjectFacilityTitle, acl.ActionWrite) {
		return nil
	}
	if !validateTitleFacility(c, facility) {
		return nil
	}
	ctx := c.Request().Context()
	cid, err := strconv.ParseInt(c.Param("cid"), 10, 32)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("invalid cid"))
	}
	titleId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("invalid title id"))
	}

	err = dbconn.Queries().DeleteUserTitle(ctx, db.DeleteUserTitleParams{
		Cid:     int32(cid),
		TitleID: titleId,
	})
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, errors.New("failed to remove title"))
	}
	return GenericSuccess(c, int(titleId))
}
