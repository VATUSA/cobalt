package endpoints

import (
	"log"
	"net/http"
	"slices"
	"strings"
	"time"
	"vatusa-cobalt/auth"
	"vatusa-cobalt/db"
	"vatusa-cobalt/models"

	"github.com/labstack/echo/v5"
)

func (h EndpointHandler) LegacySyncRoles(c *echo.Context) error {
	actor := auth.GetTokenActor(c)
	if !actor.HasGlobalGrant(auth.GrantLegacySyncRoles) {
		return echo.NewHTTPError(http.StatusForbidden, "not authorized")
	}
	ctx := c.Request().Context()
	var request models.SyncRolesRequest
	err := c.Bind(&request)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Malformed JSON")
	}

	var targetGlobalRoles []auth.UserRole
	var targetFacilityRoles []auth.FacilityRole

	for _, r := range request.Roles {
		if slices.Contains(auth.GlobalRoles, r.Role) {
			targetGlobalRoles = append(targetGlobalRoles, r.Role)
		} else if slices.Contains(auth.FacilityRoles, r.Role) {
			targetFacilityRoles = append(targetFacilityRoles, auth.FacilityRole{
				Role:     r.Role,
				Facility: r.Facility,
			})
		} else if strings.HasPrefix(r.Role, "US") {
			targetGlobalRoles = append(targetGlobalRoles, auth.RoleVATUSAStaff)
			if slices.Contains([]auth.UserRole{"US1", "US2", "US3", "US4"}, r.Role) {
				targetGlobalRoles = append(targetGlobalRoles, auth.RoleVATUSAManagement)
			}
		} else {
			// Ideally not possible, probably log something?
			log.Printf(`Role "%s" is not a valid role`, r.Role)
		}
	}

	err = h.Queries.DeleteUserGlobalRoles(ctx, int32(request.CID))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete user global roles")
	}
	err = h.Queries.DeleteUserFacilityRoles(ctx, int32(request.CID))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete user facilities roles")
	}

	for _, g := range targetGlobalRoles {
		err = h.Queries.CreateGlobalRole(ctx, db.CreateGlobalRoleParams{
			Cid:       int32(request.CID),
			Role:      g,
			CreatedAt: time.Now().Unix(),
			CreatedBy: 0,
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create global role")
		}
	}

	for _, f := range targetFacilityRoles {
		err = h.Queries.CreateFacilityRole(ctx, db.CreateFacilityRoleParams{
			Cid:       int32(request.CID),
			Role:      f.Role,
			Facility:  f.Facility,
			CreatedAt: time.Now().Unix(),
			CreatedBy: 0,
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create facility role")
		}
	}

	return c.JSON(http.StatusOK, "sync roles successful")
}
