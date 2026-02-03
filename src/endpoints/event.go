package endpoints

import (
	"net/http"
	"strconv"
	"time"
	"vatusa-cobalt/auth"
	"vatusa-cobalt/db"
	"vatusa-cobalt/models"

	"github.com/labstack/echo/v5"
)

func (h EndpointHandler) CreateEvent(c *echo.Context) error {
	ctx := c.Request().Context()
	var request models.EventRequest
	err := c.Bind(&request)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if !h.HasFacilityPermission(c, auth.PermManageEvents, request.Facility) {
		return echo.NewHTTPError(http.StatusForbidden, "missing permission")
	}

	startTime, err := time.Parse("2006-01-02 15:04", request.StartTimestamp)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid start time")
	}
	endTime, err := time.Parse("2006-01-02 15:04", request.EndTimestamp)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid end time")
	}

	err = h.Queries.CreateEvent(ctx, db.CreateEventParams{
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
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, "Event created")
}

func (h EndpointHandler) UpdateEvent(c *echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	ctx := c.Request().Context()
	var request models.EventRequest
	err = c.Bind(&request)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	event, err := h.Queries.GetEventById(ctx, int64(id))

	if event.Facility != request.Facility && !h.HasFacilityPermission(c, auth.PermManageEvents, event.Facility) {
		return echo.NewHTTPError(http.StatusForbidden, "missing permission")
	}
	if !h.HasFacilityPermission(c, auth.PermManageEvents, request.Facility) {
		return echo.NewHTTPError(http.StatusForbidden, "missing permission")
	}

	startTime, err := time.Parse("2006-01-02 15:04", request.StartTimestamp)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid start time")
	}
	endTime, err := time.Parse("2006-01-02 15:04", request.EndTimestamp)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid end time")
	}

	err = h.Queries.UpdateEvent(ctx, db.UpdateEventParams{
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
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, "Event updated")
}

func (h EndpointHandler) DeleteEvent(c *echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	ctx := c.Request().Context()
	event, err := h.Queries.GetEventById(ctx, int64(id))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if !h.HasFacilityPermission(c, auth.PermManageEvents, event.Facility) {
		return echo.NewHTTPError(http.StatusForbidden, "missing permission")
	}

	err = h.Queries.DeleteEvent(ctx, int64(id))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, "Event deleted")
}

func (h EndpointHandler) GetEventById(c *echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	ctx := c.Request().Context()
	event, err := h.Queries.GetEventById(ctx, int64(id))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	output := models.EventFromDatabase(event)
	return c.JSON(http.StatusOK, output)
}

func (h EndpointHandler) GetUpcomingEvents(c *echo.Context) error {
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
	events, err := h.Queries.GetUpcomingEvents(ctx, db.GetUpcomingEventsParams{
		StartTime: time.Now().Unix(),
		Limit:     int32(countInt),
		Offset:    int32(0),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	output := models.EventsFromDatabase(events)
	return c.JSON(http.StatusOK, output)
}

func (h EndpointHandler) GetEventsPage(c *echo.Context) error {
	page, err := strconv.Atoi(c.QueryParamOr("page", "1"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid page")
	}
	recordsPerPage := 25
	offset := (page - 1) * recordsPerPage
	ctx := c.Request().Context()

	events, err := h.Queries.GetUpcomingEvents(ctx, db.GetUpcomingEventsParams{
		StartTime: time.Now().Unix(),
		Limit:     int32(recordsPerPage),
		Offset:    int32(offset),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	output := models.EventsFromDatabase(events)
	return c.JSON(http.StatusOK, output)
}
