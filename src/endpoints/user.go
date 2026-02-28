package endpoints

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"vatusa-cobalt/acl"
	"vatusa-cobalt/models"

	"github.com/labstack/echo/v5"
)

func (h EndpointHandler) GetUser(c *echo.Context) error {
	canSeeSensitiveFields := h.HasGlobal(c, acl.ObjectUserSensitiveDetails, acl.ActionRead)
	ctx := context.Background()
	cid := c.Param("cid")
	cidInt, err := strconv.Atoi(cid)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, errors.New("invalid cid"))
	}

	user, err := h.Queries.GetCombinedUserByCID(ctx, int64(cidInt))
	if err != nil {
		return GenericError(c, http.StatusNotFound, err)
	}

	output := models.User{
		CID: cidInt,
		NetworkUser: &models.NetworkUser{
			CID:            int(user.Cid),
			FirstName:      nil,
			LastName:       nil,
			Email:          nil,
			Rating:         int(user.Rating),
			Region:         user.RegionID,
			Division:       user.DivisionID,
			SubDivision:    nil,
			PilotRating:    int(user.Pilotrating),
			MilitaryRating: int(user.Militaryrating),
		},
		DivisionUser: nil,
	}
	if user.SubdivisionID.Valid {
		output.NetworkUser.SubDivision = &user.SubdivisionID.String
	}
	if canSeeSensitiveFields {
		output.NetworkUser.FirstName = &user.NameFirst
		output.NetworkUser.LastName = &user.NameLast
		output.NetworkUser.Email = &user.Email
	}
	if user.Facility.Valid {
		output.DivisionUser = &models.DivisionUser{
			DisplayName:            nil,
			ControllerRating:       nil,
			InstructorRating:       nil,
			Facility:               user.Facility.String,
			VisitingFacilities:     []string{},
			DiscordId:              nil,
			LastPromotionTimestamp: nil,
			LastTransferTimestamp:  nil,
		}
	}
	if user.DisplayName.Valid {
		output.NetworkUser.FirstName = &user.DisplayName.String
	}
	if user.ControllerRating.Valid {
		controllerRating := int(user.ControllerRating.Int32)
		output.DivisionUser.ControllerRating = &controllerRating
	}
	if user.InstructorRating.Valid {
		instructorRating := int(user.InstructorRating.Int32)
		output.DivisionUser.InstructorRating = &instructorRating
	}
	if user.VisitingFacilities.Valid {
		visitingFacilities := strings.Split(user.VisitingFacilities.String, ",")
		output.DivisionUser.VisitingFacilities = visitingFacilities
	}
	if user.DiscordID.Valid {
		output.DivisionUser.DiscordId = &user.DiscordID.String
	}
	if user.LastPromotionTime.Valid {
		t := user.LastPromotionTime.Time.Unix()
		output.DivisionUser.LastPromotionTimestamp = &t
	}
	if user.LastTransferTime.Valid {
		t := user.LastTransferTime.Time.Unix()
		output.DivisionUser.LastTransferTimestamp = &t
	}

	return c.JSON(http.StatusOK, output)
}
