package endpoints

import (
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

	// Need to check both event.Facility and request.Facility as they might be different and the user must have write access to both
	if !AssertFacility(c, event.Facility, acl.ObjectEvent, acl.ActionWrite) {
		return err
	}
	if !AssertFacility(c, request.Facility, acl.ObjectEvent, acl.ActionWrite) {
		return err
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
	event, err := dbconn.Queries().GetEventById(ctx, int64(id))
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}
	output := models.EventFromDatabase(event)
	return c.JSON(http.StatusOK, output)
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
	events, err := dbconn.Queries().GetUpcomingEvents(ctx, db.GetUpcomingEventsParams{
		StartTime: time.Now().Unix(),
		Limit:     int32(countInt),
		Offset:    int32(0),
	})
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}

	output := models.EventsFromDatabase(events)
	return c.JSON(http.StatusOK, output)
}

func GetEventsPage(c *echo.Context) error {
	page, err := strconv.Atoi(c.QueryParamOr("page", "1"))
	if err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("invalid page"))
	}
	recordsPerPage := 25
	offset := (page - 1) * recordsPerPage
	ctx := c.Request().Context()

	events, err := dbconn.Queries().GetUpcomingEvents(ctx, db.GetUpcomingEventsParams{
		StartTime: time.Now().Unix(),
		Limit:     int32(recordsPerPage),
		Offset:    int32(offset),
	})
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}

	output := models.EventsFromDatabase(events)
	return c.JSON(http.StatusOK, output)
}
