// Package api2 holds VATSIM API v2 calls that packages imported by vatsim (e.g. roster)
// need, without creating an import cycle back into vatsim.
package api2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"vatusa-cobalt/config"
)

// UpdateMemberSubdivision sets a member's subdivision on VATSIM; nil clears it.
// https://vatsim.dev/api/core-api/members-api-update-member-details/
func UpdateMemberSubdivision(cid int64, subdivisionID *string) error {
	payload, err := json.Marshal(map[string]any{
		"id":             cid,
		"subdivision_id": subdivisionID,
	})
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/members/%d", config.VatsimApi2URL(), cid)
	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "vatusa/cobalt")
	req.Header.Set("X-API-Key", config.VatsimApi2Key())
	req.Header.Set("x-identifier", config.VatsimApi2Identifier())

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("VATSIM PATCH member %d: status %d: %s", cid, res.StatusCode, body)
	}
	return nil
}
