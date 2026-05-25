package endpoints

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"time"
	"vatusa-cobalt/acl"
	"vatusa-cobalt/auth"
	"vatusa-cobalt/config"
	"vatusa-cobalt/db"
	"vatusa-cobalt/dbconn"
	"vatusa-cobalt/models"

	"github.com/labstack/echo/v5"
)

func CreateEvent(c *echo.Context) error {
	ctx := c.Request().Context()
	var request models.EventRequest
	err := c.Bind(&request)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}

	if !AssertFacility(c, request.Facility, acl.ObjectEvent, acl.ActionWrite) {
		return nil
	}

	startTime, err := time.Parse(config.TimestampFormat, request.StartTimestamp)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("invalid start time"))
	}
	endTime, err := time.Parse(config.TimestampFormat, request.EndTimestamp)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("invalid end time"))
	}

	result, err := dbconn.Queries().CreateEvent(ctx, db.CreateEventParams{
		Title:          request.Title,
		Body:           request.Body,
		BannerImageUrl: request.BannerImageURL,
		Facility:       request.Facility,
		StartTime:      startTime.Unix(),
		EndTime:        endTime.Unix(),
		CreatedAt:      time.Now().Unix(),
		CreatedBy:      int32(auth.GetUserCid(c)),
		UpdatedAt:      time.Now().Unix(),
		UpdatedBy:      int32(auth.GetUserCid(c)),
	})
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}
	insertId, err := result.LastInsertId()
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}
	return GenericSuccess(c, int(insertId))
}

func UpdateEvent(c *echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("invalid event id"))
	}
	ctx := c.Request().Context()
	var request models.EventRequest
	err = c.Bind(&request)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}

	event, err := dbconn.Queries().GetEventById(ctx, int64(id))
	if err != nil {
		return GenericError(c, http.StatusNotFound, errors.New("event not found"))
	}

	// Check both facilities: user must have write access to both current and new facility
	if !AssertFacility(c, event.Facility, acl.ObjectEvent, acl.ActionWrite) {
		return nil
	}
	if !AssertFacility(c, request.Facility, acl.ObjectEvent, acl.ActionWrite) {
		return nil
	}

	startTime, err := time.Parse(config.TimestampFormat, request.StartTimestamp)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("invalid start time"))
	}
	endTime, err := time.Parse(config.TimestampFormat, request.EndTimestamp)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("invalid end time"))
	}

	err = dbconn.Queries().UpdateEvent(ctx, db.UpdateEventParams{
		Title:          request.Title,
		Body:           request.Body,
		BannerImageUrl: request.BannerImageURL,
		Facility:       request.Facility,
		StartTime:      startTime.Unix(),
		EndTime:        endTime.Unix(),
		UpdatedAt:      time.Now().Unix(),
		UpdatedBy:      int32(auth.GetUserCid(c)),
		ID:             int64(id),
	})
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}
	return GenericSuccess(c, id)
}

func ReviewEvent(c *echo.Context) error {
	if !AssertGlobal(c, acl.ObjectEventApproval, acl.ActionWrite) {
		return nil
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("invalid event id"))
	}
	var request models.EventReviewRequest
	if err = c.Bind(&request); err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}
	if request.Status != "approved" && request.Status != "rejected" {
		return GenericError(c, http.StatusBadRequest, errors.New("status must be 'approved' or 'rejected'"))
	}
	ctx := c.Request().Context()
	_, err = dbconn.Queries().GetEventById(ctx, int64(id))
	if err != nil {
		return GenericError(c, http.StatusNotFound, errors.New("event not found"))
	}
	reviewerCid := int32(auth.GetUserCid(c))
	err = dbconn.Queries().SetEventReviewStatus(ctx, db.SetEventReviewStatusParams{
		ReviewStatus: sql.NullString{String: request.Status, Valid: true},
		ReviewedBy:   sql.NullInt32{Int32: reviewerCid, Valid: true},
		ReviewedOn:   sql.NullInt64{Int64: time.Now().Unix(), Valid: true},
		ID:           int64(id),
	})
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}
	return GenericSuccess(c, id)
}

func DeleteEvent(c *echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("invalid event id"))
	}
	ctx := c.Request().Context()
	event, err := dbconn.Queries().GetEventById(ctx, int64(id))
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}

	if !AssertFacility(c, event.Facility, acl.ObjectEvent, acl.ActionWrite) {
		return nil
	}

	err = dbconn.Queries().DeleteEvent(ctx, int64(id))
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}
	return GenericSuccess(c, id)
}

func GetEventById(c *echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("invalid event id"))
	}
	ctx := c.Request().Context()
	var event db.Event
	if HasGlobal(c, acl.ObjectEventApproval, acl.ActionRead) {
		event, err = dbconn.Queries().GetEventById(ctx, int64(id))
	} else {
		event, err = dbconn.Queries().GetEventByIdApproved(ctx, int64(id))
	}
	if err != nil {
		return GenericError(c, http.StatusNotFound, errors.New("event not found"))
	}
	return c.JSON(http.StatusOK, models.EventFromDatabase(event))
}

func GetUpcomingEvents(c *echo.Context) error {
	count := c.Param("count")
	countInt, err := strconv.Atoi(count)
	if err != nil {
		countInt = 20
	} else if countInt < 1 {
		countInt = 1
	} else if countInt > 100 {
		countInt = 100
	}
	ctx := c.Request().Context()
	events, err := dbconn.Queries().GetUpcomingEventsApproved(ctx, db.GetUpcomingEventsApprovedParams{
		StartTime: time.Now().Unix(),
		Limit:     int32(countInt),
		Offset:    0,
	})
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, models.EventsFromDatabase(events))
}

func GetEventsPage(c *echo.Context) error {
	page, err := strconv.Atoi(c.QueryParamOr("page", "1"))
	if err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("invalid page"))
	}
	recordsPerPage := 25
	offset := (page - 1) * recordsPerPage
	ctx := c.Request().Context()
	events, err := dbconn.Queries().GetUpcomingEventsAll(ctx, db.GetUpcomingEventsAllParams{
		StartTime: time.Now().Unix(),
		Limit:     int32(recordsPerPage),
		Offset:    int32(offset),
	})
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, models.EventsFromDatabase(events))
}
