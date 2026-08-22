package legacy_migration

import (
	"context"
	"slices"
	"strings"
	"vatusa-cobalt/acl"
	"vatusa-cobalt/db"
	"vatusa-cobalt/dbconn"
	"vatusa-cobalt/models"
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

func SyncRolesForUser(request models.SyncRolesRequest) error {
	ctx := context.Background()
	modernRoles := LegacyToModernRoles(request.Roles)

	existingRoles, err := dbconn.Queries().GetRolesForUser(ctx, int32(request.CID))
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
		err = dbconn.Queries().AddRoleToUser(ctx, db.AddRoleToUserParams{
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
		err = dbconn.Queries().RemoveRoleFromUser(ctx, db.RemoveRoleFromUserParams{
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

func resolveLegacyTitleID(titleIDs map[string]map[string]int64, facility, role string) (int64, bool) {
	titleID, ok := titleIDs[facility][role]
	if !ok && strings.HasPrefix(role, "US") {
		titleID, ok = titleIDs["ZHQ"][role]
	}
	return titleID, ok
}

func BulkMigrateTitles() error {
	queries := dbconn.Queries()
	ctx := context.Background()

	facilityTitles, err := queries.GetAllFacilityTitles(ctx)
	if err != nil {
		return err
	}
	titleIDs := make(map[string]map[string]int64)
	for _, ft := range facilityTitles {
		if titleIDs[ft.Facility] == nil {
			titleIDs[ft.Facility] = make(map[string]int64)
		}
		titleIDs[ft.Facility][ft.Code] = ft.ID
	}

	assignments, err := queries.GetLegacyRoleAssignments(ctx)
	if err != nil {
		return err
	}

	for _, a := range assignments {
		titleID, ok := resolveLegacyTitleID(titleIDs, a.Facility, a.Role)
		if !ok {
			continue
		}

		err = queries.AssignUserTitle(ctx, db.AssignUserTitleParams{
			Cid:        int32(a.Cid),
			TitleID:    titleID,
			GrantorCid: 0,
			GrantedAt:  a.CreatedAt.Unix(),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func BulkMigrateRoles() error {
	queries := dbconn.Queries()
	ctx := context.Background()

	roles, err := queries.GetLegacyRoles(ctx)
	if err != nil {
		return err
	}
	var requestMap = make(map[int]*models.SyncRolesRequest)
	for _, role := range roles {
		request, ok := requestMap[int(role.Cid)]
		if !ok {
			request = &models.SyncRolesRequest{
				CID:   int(role.Cid),
				Roles: make([]models.LegacyRole, 0),
			}
			requestMap[int(role.Cid)] = request
		}
		request.Roles = append(request.Roles, models.LegacyRole{
			Role:     role.Role,
			Facility: role.Facility,
		})
	}

	for _, request := range requestMap {
		err = SyncRolesForUser(*request)
		if err != nil {
			return err
		}
	}

	return nil
}
