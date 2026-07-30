package endpoints

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"vatusa-cobalt/acl"
	"vatusa-cobalt/auth"
	"vatusa-cobalt/background"
	"vatusa-cobalt/config"
	"vatusa-cobalt/dbconn"
	"vatusa-cobalt/login"
	"vatusa-cobalt/models"
	"vatusa-cobalt/vatsim"

	"github.com/labstack/echo/v5"
)

// validatedRedirect returns the redirect query param if present and it
// passes config.IsAllowedRedirect, else "".
func validatedRedirect(c *echo.Context) string {
	redirect := c.QueryParam("redirect")
	if redirect != "" && config.IsAllowedRedirect(redirect) {
		return redirect
	}
	return ""
}

// stagingRelayPrefix marks an OAuth `state` value as belonging to a
// staging-relay login (one that GetLoginForStaging started on behalf of a
// dev/staging instance) rather than a plain login on this instance.
//
// The marker has to be a prefix rather than a standalone sentinel because
// both kinds of login can carry a caller-supplied redirect, and Connect()
// otherwise cannot tell them apart: a prod login from the legacy site sends
// state=https://www.vatusa.net/legacy/auth/callback, which is
// indistinguishable from a relay carrying the same redirect. Treating every
// non-empty state on prod as a relay sent prod logins on a detour through
// the dev instance's /token/:cid, which 500s for any user without a row in
// the dev database.
//
// A caller redirect always passes config.IsAllowedRedirect and so is always
// an absolute https URL, which can never begin with this prefix — so the
// two cases stay unambiguous.
const stagingRelayPrefix = "_staging_relay|"

// stagingRelayState builds the OAuth state for a relayed login. redirect may
// be empty, which is the default (no-redirect) relay used by webapps.
func stagingRelayState(redirect string) string {
	return stagingRelayPrefix + redirect
}

// parseStagingRelayState reports whether state marks a relay in progress and,
// if so, returns the caller redirect it carries (empty when none was given).
func parseStagingRelayState(state string) (redirect string, isRelay bool) {
	if !strings.HasPrefix(state, stagingRelayPrefix) {
		return "", false
	}
	return strings.TrimPrefix(state, stagingRelayPrefix), true
}

func GetLogin(c *echo.Context) error {
	redirect := validatedRedirect(c)
	if config.IsStaging() {
		target := "https://cobalt.vatusa.net/login/staging"
		if redirect != "" {
			target += "?redirect=" + url.QueryEscape(redirect)
		}
		return c.Redirect(http.StatusFound, target)
	}
	return c.Redirect(http.StatusFound, vatsim.ConnectFullURL(redirect))
}

func GetLogout(c *echo.Context) error {
	c.SetCookie(auth.ClearSessionCookie())
	return c.Redirect(http.StatusFound, config.PostLoginURL())
}

func Connect(c *echo.Context) error {
	code := c.QueryParam("code")
	token, err := vatsim.FetchToken(code)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "error fetching connect access token")
	}
	userData, err := vatsim.FetchUserData(token)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "error fetching connect user data")
	}
	cid, err := strconv.Atoi(userData.CID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "error extracting cid")
	}

	err = vatsim.StoreVatsimUserRecordConnect(userData)
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, errors.New("error storing vatsim user record"), err)
	}

	if userData.Vatsim.Rating.Id == config.RatingInactive || userData.Vatsim.Rating.Id == config.RatingSuspended {
		return GenericError(c, http.StatusForbidden, errors.New("account is inactive or suspended"))
	}

	if config.IsProduction() || config.IsStaging() {
		job := background.NewJob("vatsim_sync", fmt.Sprintf("%d", cid))
		err = job.Run()
		if err != nil {
			log.Printf("error syncing user cid %d: %v", cid, err)
		}
	}

	user, err := dbconn.GetCombinedUserByCID(cid)
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}
	if user == nil {
		return GenericError(c, http.StatusInternalServerError, errors.New("user record does not exist"))
	}

	jwt, err := login.CreateTokenForUser(*user)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create token")
	}
	c.SetCookie(auth.NewSessionCookie(jwt))

	// state carries a caller-supplied post-login redirect target across the
	// VATSIM Connect round trip (see validatedRedirect / GetLogin), and may
	// additionally be tagged as a staging relay (see stagingRelayPrefix).
	// Because VATSIM's redirect_uri is only registered for cobalt.vatusa.net,
	// this handler runs on prod both for prod's own logins and for dev/staging
	// logins proxied via GetLoginForStaging's relay.
	//
	// Only a relay-tagged state means a handoff is in progress: loop back into
	// /login/staging (now logged in) to finish it. An untagged state is an
	// ordinary login on this instance that asked for its own redirect target,
	// so honor it directly — sending those into /login/staging would relay a
	// prod login through the dev instance.
	if state := c.QueryParam("state"); state != "" {
		if relayRedirect, isRelay := parseStagingRelayState(state); isRelay && config.IsProduction() {
			target := "/login/staging"
			if relayRedirect != "" && config.IsAllowedRedirect(relayRedirect) {
				target += "?redirect=" + url.QueryEscape(relayRedirect)
			}
			return c.Redirect(http.StatusFound, target)
		}
		if config.IsAllowedRedirect(state) {
			return c.Redirect(http.StatusFound, state)
		}
	}

	return c.Redirect(http.StatusFound, config.PostLoginURL())
}

func GetGenerateUserToken(c *echo.Context) error {
	if !AssertGlobal(c, acl.ObjectLegacyLoginToken, acl.ActionWrite) {
		return nil
	}
	cid, err := strconv.Atoi(c.Param("cid"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid cid")
	}
	user, err := dbconn.GetCombinedUserByCID(cid)
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}
	if user == nil {
		return GenericError(c, http.StatusInternalServerError, errors.New("user record does not exist"))
	}
	token, err := login.CreateTokenForUser(*user)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create token")
	}
	data := make(map[string]string)
	data["token"] = token
	return c.JSON(http.StatusOK, data)
}

func PostUserDetailsFromToken(c *echo.Context) error {
	if !AssertGlobal(c, acl.ObjectLegacyLoginToken, acl.ActionRead) {
		return nil
	}
	var request models.TokenSessionRequest
	err := c.Bind(&request)
	if err != nil {
		return GenericError(c, http.StatusBadRequest, err)
	}
	cid, err := auth.GetCIDFromToken(request.Token)
	if err != nil {
		return GenericError(c, http.StatusUnauthorized, errors.New("invalid token"))
	}
	user, err := dbconn.GetCombinedUserByCID(cid)
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}
	if user == nil {
		return c.JSON(http.StatusNotFound, models.Session{})
	}

	userModel := models.UserFromDatabase(*user, true, nil)

	permissionHandler := GetPermissionHandler(c)

	globalPermissions := permissionHandler.GetGlobalPermissions()
	facilityPermissions := permissionHandler.GetFacilityPermissions()

	session := models.Session{
		User:                &userModel,
		GlobalPermissions:   []models.GlobalPermission{},
		FacilityPermissions: []models.FacilityPermission{},
	}

	for _, permission := range globalPermissions {
		session.GlobalPermissions = append(session.GlobalPermissions, models.GlobalPermission{
			Action: string(permission.Action),
			Object: string(permission.Object),
		})
	}

	for _, permission := range facilityPermissions {
		session.FacilityPermissions = append(session.FacilityPermissions, models.FacilityPermission{
			Action:   string(permission.Action),
			Object:   string(permission.Object),
			Facility: permission.Facility,
		})
	}

	return c.JSON(http.StatusOK, session)
}

func LoginAs(c *echo.Context) error {
	if !config.IsDevelopment() {
		return echo.NewHTTPError(http.StatusNotFound, "Not found")
	}
	cid, err := strconv.Atoi(c.Param("cid"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad Request")
	}
	user, err := dbconn.GetCombinedUserByCID(cid)
	if err != nil {
		return GenericError(c, http.StatusInternalServerError, err)
	}
	if user == nil {
		return GenericError(c, http.StatusInternalServerError, errors.New("user record does not exist"))
	}

	token, err := login.CreateTokenForUser(*user)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create token")
	}
	c.SetCookie(auth.NewSessionCookie(token))
	return c.JSON(http.StatusOK, "success")
}

func WhoAmI(c *echo.Context) error {
	cid := auth.GetUserCid(c)

	return c.String(http.StatusOK, fmt.Sprintf("%d", cid))
}

// stagingTokenError logs why the prod->staging /token/:cid relay failed and
// returns the generic 500 shown to the browser. The cause is logged rather
// than returned because it can carry internal service detail.
func stagingTokenError(cid int, cause error) error {
	log.Printf("staging relay: failed to mint token for cid %d via %s: %v", cid, config.StagingInternalURL(), cause)
	return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate staging token")
}

func GetLoginForStaging(c *echo.Context) error {
	if !config.IsProduction() {
		return echo.NewHTTPError(http.StatusNotFound, "Not found")
	}
	redirect := validatedRedirect(c)
	if !auth.IsLoggedIn(c) {
		return c.Redirect(http.StatusFound, vatsim.ConnectFullURL(stagingRelayState(redirect)))
	}
	cid := auth.GetUserCid(c)

	client := &http.Client{}
	tokenURL := fmt.Sprintf("%s/token/%d", config.StagingInternalURL(), cid)
	req, err := http.NewRequest("GET", tokenURL, nil)
	if err != nil {
		return stagingTokenError(cid, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(auth.ActorTokenHeader, config.StagingActorToken())

	resp, err := client.Do(req)
	if err != nil {
		return stagingTokenError(cid, err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return stagingTokenError(cid, err)
	}

	// The staging instance answers with {"token": "..."} on success and a
	// models.GenericResponse on failure, whose `errors` is an array — so this
	// deliberately does not decode into map[string]string, which fails to
	// unmarshal the error shape at all and hides why the relay broke.
	var data struct {
		Token  string   `json:"token"`
		Errors []string `json:"errors"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return stagingTokenError(cid, fmt.Errorf("status %d, unparseable body %q", resp.StatusCode, string(body)))
	}
	if resp.StatusCode != http.StatusOK || data.Token == "" {
		return stagingTokenError(cid, fmt.Errorf("status %d: %s", resp.StatusCode, strings.Join(data.Errors, "; ")))
	}
	redirectUrl := fmt.Sprintf("%s/login/useToken/%s", config.StagingPublicURL(), data.Token)
	if redirect != "" {
		redirectUrl += "?redirect=" + url.QueryEscape(redirect)
	}
	return c.Redirect(http.StatusFound, redirectUrl)
}

func LoginUseToken(c *echo.Context) error {
	if !config.IsStaging() {
		return echo.NewHTTPError(http.StatusNotFound, "Not found")
	}
	token := c.Param("token")

	c.SetCookie(auth.NewSessionCookie(token))

	if redirect := validatedRedirect(c); redirect != "" {
		return c.Redirect(http.StatusFound, redirect)
	}
	return c.Redirect(http.StatusFound, config.PostLoginURL())
}
