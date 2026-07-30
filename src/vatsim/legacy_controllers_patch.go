// LEGACY-CONTROLLERS-PATCH: temporary support for `current` (legacy Laravel site), which
// still authenticates via `vatusa-old.controllers`. Delete this file, its sqlc query
// (sql/queries/legacy_controllers_patch.sql), and the call site in vatsim_user_hooks.go
// once `current` no longer depends on that table.
package vatsim

import (
	"context"
	"log"
	"vatusa-cobalt/config"
	"vatusa-cobalt/db"
	"vatusa-cobalt/dbconn"
)

func provisionLegacyController(cid int64, vatsimUser db.VatsimUser, facility string) {
	homeController := int32(0)
	if facility == config.FacilityAcademy {
		homeController = 1
	}
	err := dbconn.Queries().ProvisionLegacyController(context.Background(), db.ProvisionLegacyControllerParams{
		Cid:                uint32(cid),
		Fname:              vatsimUser.NameFirst,
		Lname:              vatsimUser.NameLast,
		Email:              vatsimUser.Email,
		Facility:           facility,
		Rating:             vatsimUser.Rating,
		FlagHomecontroller: homeController,
	})
	if err != nil {
		// Best-effort: vatusa-old is being phased out and must never block a
		// cobalt/webapps login. Log and move on.
		log.Printf("LEGACY-CONTROLLERS-PATCH: failed to provision vatusa-old.controllers for cid %d: %v", cid, err)
	}
}
