package endpoints

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
	"vatusa-cobalt/acl"
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
		return time.Time{}, errors.New("expires must not be more than 45 days from today")
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
	if position == "" {
		return GenericError(c, http.StatusBadRequest, errors.New("position is required"))
	}
	expires, err := parseSoloCertExpires(request.Expires)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}

	facility, ok := AssertFacilityForCid(c, request.Cid, acl.ObjectSoloCert, acl.ActionWrite)
	if !ok {
		return nil
	}

	now := time.Now()
	result, err := dbconn.Queries().CreateSoloCert(ctx, db.CreateSoloCertParams{
		Cid:       int64(request.Cid),
		Facility:  facility,
		Position:  position,
		Expires:   expires,
		CreatedAt: now,
		UpdatedAt: now,
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
	certId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("invalid solo cert id"))
	}
	ctx := c.Request().Context()

	existing, err := dbconn.Queries().GetSoloCertById(ctx, int64(certId))
	if err != nil {
		return GenericError(c, http.StatusNotFound, errors.New("solo cert not found"))
	}

	var request models.SoloCertUpdateRequest
	err = c.Bind(&request)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}
	position := strings.TrimSpace(request.Position)
	if position == "" {
		return GenericError(c, http.StatusBadRequest, errors.New("position is required"))
	}
	expires, err := parseSoloCertExpires(request.Expires)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}

	facility, ok := AssertFacilityForCid(c, int(existing.Cid), acl.ObjectSoloCert, acl.ActionWrite)
	if !ok {
		return nil
	}

	err = dbconn.Queries().UpdateSoloCert(ctx, db.UpdateSoloCertParams{
		Facility:  facility,
		Position:  position,
		Expires:   expires,
		UpdatedAt: time.Now(),
		ID:        int64(certId),
	})
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, errors.New("failed to update solo cert"))
	}
	return GenericSuccess(c, certId)
}

func DeleteSoloCert(c *echo.Context) error {
	certId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("invalid solo cert id"))
	}
	ctx := c.Request().Context()

	existing, err := dbconn.Queries().GetSoloCertById(ctx, int64(certId))
	if err != nil {
		return GenericError(c, http.StatusNotFound, errors.New("solo cert not found"))
	}

	if _, ok := AssertFacilityForCid(c, int(existing.Cid), acl.ObjectSoloCert, acl.ActionWrite); !ok {
		return nil
	}

	err = dbconn.Queries().DeleteSoloCert(ctx, int64(certId))
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, errors.New("failed to delete solo cert"))
	}
	return GenericSuccess(c, certId)
}
