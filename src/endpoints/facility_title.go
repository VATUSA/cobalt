package endpoints

import (
	"context"
	"database/sql"
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
		if errors.Is(err, sql.ErrNoRows) {
			_ = RespondError(c, http.StatusBadRequest, errors.New("unknown facility"))
		} else {
			_ = RespondError(c, http.StatusInternalServerError, err)
		}
		return false
	}
	return true
}

func userHoldsTitle(ctx context.Context, cid int64, facility string, titleId int64) (bool, error) {
	titles, err := dbconn.Queries().GetUserTitlesByFacility(ctx, db.GetUserTitlesByFacilityParams{
		Cid:      cid,
		Facility: facility,
	})
	if err != nil {
		return false, err
	}
	for _, t := range titles {
		if t.ID == titleId {
			return true, nil
		}
	}
	return false, nil
}

func resolveTitleTierPermission(c *echo.Context, facility string, tier string) bool {
	permissionObject, ok := acl.TitleTierToPermissionObjectMap[acl.TitleTier(tier)]
	if !ok {
		_ = RespondError(c, http.StatusBadRequest, errors.New("title tier cannot be assigned"))
		return false
	}
	return AssertFacility(c, facility, permissionObject, acl.ActionWrite)
}

func resolveActor(c *echo.Context) (int64, *db.GetCombinedUserRow, bool) {
	actorCid := int64(auth.GetUserCid(c))
	actor, err := dbconn.GetCombinedUserByCID(int(actorCid))
	if err != nil {
		_ = RespondError(c, http.StatusInternalServerError, err)
		return 0, nil, false
	}
	if actor == nil {
		_ = RespondError(c, http.StatusInternalServerError, errors.New("actor not found"))
		return 0, nil, false
	}
	return actorCid, actor, true
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
	if _, ok := acl.TitleTierToPermissionObjectMap[acl.TitleTier(request.Tier)]; !ok {
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
		Cid:      cid,
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
	if !resolveTitleTierPermission(c, facility, title.Tier) {
		return nil
	}

	user, err := dbconn.GetCombinedUserByCID(int(cid))
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, err)
	}
	if user == nil {
		return RespondError(c, http.StatusNotFound, errors.New("user not found"))
	}

	held, err := userHoldsTitle(ctx, cid, facility, titleId)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, err)
	}
	if held {
		return RespondError(c, http.StatusConflict, errors.New("title already assigned"))
	}

	actorCid, actor, ok := resolveActor(c)
	if !ok {
		return nil
	}

	err = dbconn.WithTransaction(ctx, func(q *db.Queries) error {
		err := q.AssignUserTitle(ctx, db.AssignUserTitleParams{
			Cid:        cid,
			TitleID:    titleId,
			GrantorCid: int32(actorCid),
			GrantedAt:  time.Now().Unix(),
		})
		if err != nil {
			return err
		}
		return action.Log(q, *user, action.TitleGranted,
			fmt.Sprintf("Granted title %s at %s by %s (%d)", title.Title, facility, actor.DisplayName, actorCid), actorCid)
	})
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, err)
	}

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
	if !resolveTitleTierPermission(c, facility, title.Tier) {
		return nil
	}

	user, err := dbconn.GetCombinedUserByCID(int(cid))
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, err)
	}
	if user == nil {
		return RespondError(c, http.StatusNotFound, errors.New("user not found"))
	}

	held, err := userHoldsTitle(ctx, cid, facility, titleId)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, err)
	}
	if !held {
		return RespondError(c, http.StatusNotFound, errors.New("title not found"))
	}

	actorCid, actor, ok := resolveActor(c)
	if !ok {
		return nil
	}

	err = dbconn.WithTransaction(ctx, func(q *db.Queries) error {
		err := q.DeleteUserTitle(ctx, db.DeleteUserTitleParams{
			Cid:     cid,
			TitleID: titleId,
		})
		if err != nil {
			return err
		}
		return action.Log(q, *user, action.TitleRevoked,
			fmt.Sprintf("Revoked title %s at %s by %s (%d)", title.Title, facility, actor.DisplayName, actorCid), actorCid)
	})
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, err)
	}

	return RespondSuccess(c, int(titleId))
}
