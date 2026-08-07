package endpoints

import (
	"context"
	"database/sql"
	"errors"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"
	"vatusa-cobalt/acl"
	"vatusa-cobalt/auth"
	"vatusa-cobalt/config"
	"vatusa-cobalt/db"
	"vatusa-cobalt/dbconn"
	"vatusa-cobalt/models"
	"vatusa-cobalt/storage"

	"github.com/labstack/echo/v5"
)

// bannerFormField is the multipart field the staff app posts the chosen banner
// image under. Events used to carry a caller-supplied banner_image_url, which
// in practice meant Imgur or Discord links; we now host the image ourselves and
// generate the URL, so the field is still what lands in the database but its
// value comes from us.
const bannerFormField = "banner_image"

// bindEventRequest fills request from the incoming body and returns the
// uploaded banner, if any. Multipart submissions come from the staff app and
// may carry a file; a JSON body is still accepted so API clients that already
// have a hosted URL keep working. In both cases banner_image_url is honoured
// as the existing value, which is what lets an edit that doesn't replace the
// image keep the banner it already has.
func bindEventRequest(c *echo.Context, request *models.EventRequest) (*multipart.FileHeader, error) {
	contentType := c.Request().Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		return nil, c.Bind(request)
	}

	// ParseMultipartForm will happily spill an unbounded body to temp files, so
	// cap the whole request first. The slack over MaxBannerBytes covers the
	// multipart framing and the other form fields.
	req := c.Request()
	req.Body = http.MaxBytesReader(c.Response(), req.Body, storage.MaxBannerBytes+(1<<20))

	// Small in-memory budget: anything bigger spills to a temp file that Go
	// removes when the request ends, rather than being held in the heap.
	if err := req.ParseMultipartForm(1 << 20); err != nil {
		return nil, errors.New("could not read the submitted form (is the image too large?)")
	}

	form := req.MultipartForm
	field := func(name string) string {
		if values := form.Value[name]; len(values) > 0 {
			return strings.TrimSpace(values[0])
		}
		return ""
	}

	request.Title = field("title")
	request.Body = field("body")
	request.Facility = field("facility")
	request.BannerImageURL = field("banner_image_url")
	request.StartTimestamp = field("start_timestamp")
	request.EndTimestamp = field("end_timestamp")

	if files := form.File[bannerFormField]; len(files) > 0 && files[0].Size > 0 {
		return files[0], nil
	}
	return nil, nil
}

// resolveBannerURL uploads a submitted banner and returns the URL to store, or
// falls back to the URL already on the request when no new file was chosen.
// Call it only after the facility permission check has passed, so an
// unauthorised request can't write to the bucket.
func resolveBannerURL(ctx context.Context, facility, existingURL string, banner *multipart.FileHeader) (string, int, error) {
	if banner == nil {
		if existingURL == "" {
			return "", http.StatusBadRequest, errors.New("a banner image is required")
		}
		return existingURL, 0, nil
	}

	url, err := storage.UploadEventBanner(ctx, facility, banner)
	if err != nil {
		if errors.Is(err, storage.ErrInvalidImage) {
			return "", http.StatusBadRequest, err
		}
		return "", http.StatusInternalServerError, err
	}
	return url, 0, nil
}

func CreateEvent(c *echo.Context) error {
	ctx := c.Request().Context()
	var request models.EventRequest
	banner, err := bindEventRequest(c, &request)
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

	// Last thing before the insert: an upload that succeeds but is then
	// followed by a rejected request leaves an orphaned object in the bucket.
	bannerURL, status, err := resolveBannerURL(ctx, request.Facility, request.BannerImageURL, banner)
	if err != nil {
		return GenericError(c, status, err)
	}

	result, err := dbconn.Queries().CreateEvent(ctx, db.CreateEventParams{
		Title:          request.Title,
		Body:           request.Body,
		BannerImageUrl: bannerURL,
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
	banner, err := bindEventRequest(c, &request)
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

	// Without a replacement file the event keeps the banner it already has,
	// even if the client omitted banner_image_url entirely. Left until last so
	// a request rejected above doesn't orphan an object in the bucket.
	existingBannerURL := request.BannerImageURL
	if existingBannerURL == "" {
		existingBannerURL = event.BannerImageUrl
	}
	bannerURL, status, err := resolveBannerURL(ctx, request.Facility, existingBannerURL, banner)
	if err != nil {
		return GenericError(c, status, err)
	}

	err = dbconn.Queries().UpdateEvent(ctx, db.UpdateEventParams{
		Title:          request.Title,
		Body:           request.Body,
		BannerImageUrl: bannerURL,
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
	event, err := dbconn.Queries().GetEventById(ctx, int64(id))
	if err != nil {
		return GenericError(c, http.StatusNotFound, errors.New("event not found"))
	}
	if event.ReviewStatus.String != "approved" {
		authorized := HasGlobal(c, acl.ObjectEventApproval, acl.ActionRead) ||
			HasFacility(c, event.Facility, acl.ObjectEvent, acl.ActionWrite)
		if !authorized {
			return GenericError(c, http.StatusNotFound, errors.New("event not found"))
		}
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
		EndTime: time.Now().Unix(),
		Limit:   int32(countInt),
		Offset:  0,
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
	facility := strings.ToUpper(strings.TrimSpace(c.QueryParamOr("facility", "")))
	recordsPerPage := 25
	offset := (page - 1) * recordsPerPage
	ctx := c.Request().Context()
	var events []db.Event
	switch {
	case facility != "" && HasGlobal(c, acl.ObjectEventApproval, acl.ActionRead):
		events, err = dbconn.Queries().GetUpcomingEventsForFacilityOrPending(ctx, db.GetUpcomingEventsForFacilityOrPendingParams{
			StartTime: time.Now().Unix(),
			Facility:  facility,
			Limit:     int32(recordsPerPage),
			Offset:    int32(offset),
		})
	case facility != "" && HasFacility(c, facility, acl.ObjectEvent, acl.ActionWrite):
		events, err = dbconn.Queries().GetUpcomingEventsForFacility(ctx, db.GetUpcomingEventsForFacilityParams{
			StartTime: time.Now().Unix(),
			Facility:  facility,
			Limit:     int32(recordsPerPage),
			Offset:    int32(offset),
		})
	case facility == "" && HasGlobal(c, acl.ObjectEventApproval, acl.ActionRead):
		events, err = dbconn.Queries().GetUpcomingEventsAll(ctx, db.GetUpcomingEventsAllParams{
			StartTime: time.Now().Unix(),
			Limit:     int32(recordsPerPage),
			Offset:    int32(offset),
		})
	default:
		events, err = dbconn.Queries().GetUpcomingEventsApproved(ctx, db.GetUpcomingEventsApprovedParams{
			EndTime: time.Now().Unix(),
			Limit:   int32(recordsPerPage),
			Offset:  int32(offset),
		})
	}
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, models.EventsFromDatabase(events))
}
