package endpoints

import (
	"context"
	"net/http"
	"slices"
	"strings"
	"vatusa-cobalt/acl"
	"vatusa-cobalt/db"
	"vatusa-cobalt/models"

	"github.com/labstack/echo/v5"
)

var LegacyToModernRoleMap = map[string]acl.Role{
	"ATM":  acl.RoleAirTrafficManager,
	"DATM": acl.RoleDeputyAirTrafficManager,
	"TA":   acl.RoleTrainingAdministrator,
	"EC":   acl.RoleEventCoordinator,
	"FE":   acl.RoleFacilityEngineer,
	"WM":   acl.RoleWebMaintainer,

	"INS":    acl.RoleInstructor,
	"MTR":    acl.RoleMentor,
	"CBT":    acl.RoleDivisionAcademyEditor,
	"FACCBT": acl.RoleFacilityAcademyEditor,

	"ACE":  acl.RoleACETeam,
	"DCC":  acl.RoleDivisionCommandCenter,
	"SMT":  acl.RoleSocialMediaTeam,
	"USWT": acl.RoleDivisionTechTeam,

	"EMAIL": acl.RoleEmailUser,
}

func LegacyToModernRoles(roles []models.LegacyRole) []acl.ScopedRole {
	var result []acl.ScopedRole

	for _, role := range roles {
		r, ok := LegacyToModernRoleMap[role.Role]
		if ok {
			isGlobalRole := slices.Contains(acl.GlobalRoles, r)
			fac := role.Facility
			if isGlobalRole {
				fac = acl.ScopedRoleGlobalFacility
			}
			result = append(result, acl.ScopedRole{
				Role:     r,
				Facility: fac,
			})
		} else if strings.HasPrefix(role.Role, "US") && role.Role != "USWT" {
			result = append(result, acl.ScopedRole{
				Role:     acl.RoleDivisionStaff,
				Facility: acl.ScopedRoleGlobalFacility,
			})
			if slices.Contains([]string{"US1", "US2", "US3", "US4"}, role.Role) {
				result = append(result, acl.ScopedRole{
					Role:     acl.RoleDivisionManagement,
					Facility: acl.ScopedRoleGlobalFacility,
				})
			}
			if role.Role == "US0" || role.Role == "US6" {
				result = append(result, acl.ScopedRole{
					Role:     acl.RoleSystemAdmin,
					Facility: acl.ScopedRoleGlobalFacility,
				})
			}
		}
	}

	return result
}

func DetermineDropAddRoles(existingRoles []acl.ScopedRole, targetRoles []acl.ScopedRole) (add []acl.ScopedRole, drop []acl.ScopedRole) {
	for _, a := range existingRoles {
		found := false
		for _, b := range targetRoles {
			if a.Role == b.Role && a.Facility == b.Facility {
				found = true
				break
			}
		}
		if !found {
			drop = append(drop, a)
		}
	}

	for _, b := range targetRoles {
		found := false
		for _, a := range existingRoles {
			if a.Role == b.Role && a.Facility == b.Facility {
				found = true
			}
		}
		if !found {
			add = append(add, b)
		}
	}
	return
}

func SyncRolesForUser(ctx context.Context, queries *db.Queries, request models.SyncRolesRequest) error {
	modernRoles := LegacyToModernRoles(request.Roles)

	existingRoles, err := queries.GetRolesForUser(ctx, int32(request.CID))
	if err != nil {
		return err
	}

	var existingRolesScoped []acl.ScopedRole
	for _, role := range existingRoles {
		existingRolesScoped = append(existingRolesScoped, acl.ScopedRole{
			Role:     acl.Role(role.Role),
			Facility: role.Facility,
		})
	}

	addRoles, dropRoles := DetermineDropAddRoles(existingRolesScoped, modernRoles)

	for _, addRole := range addRoles {
		err = queries.AddRoleToUser(ctx, db.AddRoleToUserParams{
			Cid:        int32(request.CID),
			Facility:   addRole.Facility,
			Role:       string(addRole.Role),
			GrantorCid: 0,
			GrantedAt:  0,
		})
		if err != nil {
			return err
		}
	}

	for _, dropRole := range dropRoles {
		err = queries.RemoveRoleFromUser(ctx, db.RemoveRoleFromUserParams{
			Cid:      int32(request.CID),
			Facility: dropRole.Facility,
			Role:     string(dropRole.Role),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (h EndpointHandler) LegacySyncRoles(c *echo.Context) error {
	if !h.AssertGlobal(c, acl.ObjectLegacyRoleSync, acl.ActionWrite) {
		return nil
	}
	var request models.SyncRolesRequest
	err := c.Bind(&request)
	if err != nil {
		return err
	}
	ctx := c.Request().Context()
	err = SyncRolesForUser(ctx, h.Queries, request)
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, "sync roles successful")
}

func (h EndpointHandler) LegacySyncRolesBulk(c *echo.Context) error {

	if !h.AssertGlobal(c, acl.ObjectLegacyRoleSync, acl.ActionWrite) {
		return nil
	}
	var request models.BulkSyncRolesRequest
	err := c.Bind(&request)
	if err != nil {
		return err
	}
	ctx := c.Request().Context()
	for _, req := range request.Requests {
		err = SyncRolesForUser(ctx, h.Queries, req)
		if err != nil {
			return GenericError(c, http.StatusInternalServerError, err)
		}
	}

	return c.JSON(http.StatusOK, "sync roles successful")
}
