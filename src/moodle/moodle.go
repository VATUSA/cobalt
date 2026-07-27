// Package moodle is a minimal client for the VATUSA Academy Moodle instance's
// REST web service, covering just enough of the core_user_*/core_cohort_*
// functions to keep a controller's cohort membership (e.g. the "OBS" cohort
// for new Observer-rated controllers) in sync at login time. This replaces
// the legacy api repo's ULSHelper::doHandleLogin, which stopped firing once
// login moved to cobalt/webapps.
package moodle

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
	"vatusa-cobalt/config"
)

const requestTimeout = 30 * time.Second

var httpClient = &http.Client{Timeout: requestTimeout}

// request calls a Moodle web service function and decodes the JSON response
// into out (which may be nil if the caller doesn't need the body).
func request(function string, params url.Values, method string, out interface{}) error {
	if params == nil {
		params = url.Values{}
	}
	params.Set("wstoken", config.MoodleToken())
	params.Set("moodlewsrestformat", "json")
	params.Set("wsfunction", function)

	var resp *http.Response
	var err error
	if method == http.MethodPost {
		resp, err = httpClient.PostForm(config.MoodleURL()+"/webservice/rest/server.php", params)
	} else {
		resp, err = httpClient.Get(config.MoodleURL() + "/webservice/rest/server.php?" + params.Encode())
	}
	if err != nil {
		return fmt.Errorf("moodle: request %s: %w", function, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("moodle: reading response for %s: %w", function, err)
	}

	// Moodle reports web service errors with HTTP 200 and an {"exception":...}
	// body rather than a non-2xx status, so check the body shape regardless of
	// status code.
	var exception struct {
		Exception string `json:"exception"`
		Message   string `json:"message"`
	}
	if err := json.Unmarshal(body, &exception); err == nil && exception.Exception != "" {
		return fmt.Errorf("moodle: %s failed: %s", function, exception.Message)
	}

	if out == nil {
		return nil
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("moodle: decoding response for %s: %w", function, err)
	}
	return nil
}

// GetUserID looks up the Moodle user id for a VATSIM CID (stored as the
// Moodle user's idnumber). Returns ok=false if no such user exists yet.
func GetUserID(cid int64) (id int, ok bool, err error) {
	params := url.Values{
		"criteria[0][key]":   {"idnumber"},
		"criteria[0][value]": {strconv.FormatInt(cid, 10)},
	}
	var out struct {
		Users []struct {
			ID int `json:"id"`
		} `json:"users"`
	}
	if err := request("core_user_get_users", params, http.MethodGet, &out); err != nil {
		return 0, false, err
	}
	if len(out.Users) == 0 {
		return 0, false, nil
	}
	return out.Users[0].ID, true, nil
}

// CreateUser creates a Moodle account for a VATSIM CID and returns its new
// Moodle user id. Mirrors api's VATUSAMoodle::createUser.
func CreateUser(cid int64, firstName, lastName, email string) (int, error) {
	cidStr := strconv.FormatInt(cid, 10)
	params := url.Values{
		"users[0][createpassword]": {"0"},
		"users[0][username]":       {cidStr},
		"users[0][password]":       {randomPassword()},
		"users[0][auth]":           {"manual"},
		"users[0][firstname]":      {firstName},
		"users[0][lastname]":       {lastName},
		"users[0][email]":          {email},
		"users[0][maildisplay]":    {"0"},
		"users[0][idnumber]":       {cidStr},
		"users[0][mailformat]":     {"1"},
	}
	var out []struct {
		ID int `json:"id"`
	}
	if err := request("core_user_create_users", params, http.MethodPost, &out); err != nil {
		return 0, err
	}
	if len(out) == 0 {
		return 0, fmt.Errorf("moodle: core_user_create_users returned no users for cid %d", cid)
	}
	return out[0].ID, nil
}

// AssignCohort adds a Moodle user to a cohort identified by its idnumber
// (e.g. "OBS"). Moodle's core_cohort_add_cohort_members is a no-op if the
// user is already a member, so this is safe to call on every login.
func AssignCohort(moodleUserID int, cohortIdnumber string) error {
	params := url.Values{
		"members[0][cohorttype][type]":  {"idnumber"},
		"members[0][cohorttype][value]": {cohortIdnumber},
		"members[0][usertype][type]":    {"id"},
		"members[0][usertype][value]":   {strconv.Itoa(moodleUserID)},
	}
	return request("core_cohort_add_cohort_members", params, http.MethodPost, nil)
}

// randomPassword generates a throwaway password for accounts created via the
// web service. Moodle SSO (userkey auth) is used for actual logins, so this
// value is never entered by anyone; it only needs to satisfy Moodle's create
// call.
func randomPassword() string {
	return fmt.Sprintf("Vatusa!%d", time.Now().UnixNano())
}
