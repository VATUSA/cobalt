package endpoints

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"vatusa-cobalt/acl"
	"vatusa-cobalt/auth"
	"vatusa-cobalt/db"
	"vatusa-cobalt/dbconn"
	"vatusa-cobalt/models"

	"github.com/labstack/echo/v5"
)

const soloCertMaxExpiresDays = 45

func parseSoloCertExpires(raw string) (time.Time, error) {
	expires, err := time.Parse(time.DateOnly, raw)
	if err != nil {
		return time.Time{}, errors.New("expires must be a date in YYYY-MM-DD format")
	}

	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if expires.Before(today) {
		return time.Time{}, errors.New("expires must not be in the past")
	}
	if expires.After(today.AddDate(0, 0, soloCertMaxExpiresDays)) {
		return time.Time{}, fmt.Errorf("expires must not be more than %d days from today", soloCertMaxExpiresDays)
	}

	return expires, nil
}

func GetSoloCerts(c *echo.Context) error {
	ctx := c.Request().Context()

	certs, err := dbconn.Queries().GetActiveSoloCerts(ctx)
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, models.SoloCertsFromDatabase(certs))
}

func CreateSoloCert(c *echo.Context) error {
	ctx := c.Request().Context()
	var request models.SoloCertRequest
	err := c.Bind(&request)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}

	position := strings.TrimSpace(request.Position)
	if err := requireText("position", position, 20); err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}
	expires, err := parseSoloCertExpires(request.Expires)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}

	// The facility returned here is the one that granted access (the
	// controller's home facility, or the specific visiting facility the
	// caller matched on) — that's what gets stamped onto the record.
	facility, ok := AssertFacilityForCid(c, request.Cid, acl.ObjectSoloCert, acl.ActionWrite)
	if !ok {
		return nil
	}

	existing, err := dbconn.Queries().GetActiveSoloCertForCidPosition(ctx, db.GetActiveSoloCertForCidPositionParams{
		Cid:      int64(request.Cid),
		Position: position,
	})
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}
	if len(existing) > 0 {
		return GenericError(c, http.StatusConflict, errors.New("this controller already has an active solo cert for this position"))
	}

	userCid := int32(auth.GetUserCid(c))
	now := time.Now()
	result, err := dbconn.Queries().CreateSoloCert(ctx, db.CreateSoloCertParams{
		Cid:          int64(request.Cid),
		Facility:     facility,
		Position:     position,
		Expires:      expires,
		CreatedByCid: userCid,
		UpdatedByCid: userCid,
		CreatedAt:    now,
		UpdatedAt:    now,
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

func UpdateSoloCert(c *echo.Context) error {
	certId, ok := parseId64(c, "id")
	if !ok {
		return nil
	}
	ctx := c.Request().Context()

	existing, err := dbconn.Queries().GetSoloCertById(ctx, certId)
	if err != nil {
		return GenericError(c, http.StatusNotFound, errors.New("solo cert not found"))
	}

	// Scoped to the facility that granted the cert, not the controller's
	// current facility — a transfer must not silently move edit rights to
	// the new facility.
	if !AssertFacility(c, existing.Facility, acl.ObjectSoloCert, acl.ActionWrite) {
		return nil
	}

	var request models.SoloCertUpdateRequest
	err = c.Bind(&request)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}
	position := strings.TrimSpace(request.Position)
	if err := requireText("position", position, 20); err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}
	expires, err := parseSoloCertExpires(request.Expires)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}

	result, err := dbconn.Queries().UpdateSoloCert(ctx, db.UpdateSoloCertParams{
		Position:     position,
		Expires:      expires,
		UpdatedByCid: int32(auth.GetUserCid(c)),
		UpdatedAt:    time.Now(),
		ID:           certId,
	})
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, errors.New("failed to update solo cert"))
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return GenericError(c, http.StatusNotFound, errors.New("solo cert not found"))
	}
	return GenericSuccess(c, int(certId))
}

func DeleteSoloCert(c *echo.Context) error {
	certId, ok := parseId64(c, "id")
	if !ok {
		return nil
	}
	ctx := c.Request().Context()

	existing, err := dbconn.Queries().GetSoloCertById(ctx, certId)
	if err != nil {
		return GenericError(c, http.StatusNotFound, errors.New("solo cert not found"))
	}

	if !AssertFacility(c, existing.Facility, acl.ObjectSoloCert, acl.ActionWrite) {
		return nil
	}

	result, err := dbconn.Queries().DeleteSoloCert(ctx, certId)
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, errors.New("failed to delete solo cert"))
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return GenericError(c, http.StatusNotFound, errors.New("solo cert not found"))
	}
	return GenericSuccess(c, int(certId))
}
