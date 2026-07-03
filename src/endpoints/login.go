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
	// VATSIM Connect round trip (see validatedRedirect / GetLogin). Because
	// VATSIM's redirect_uri is only registered for cobalt.vatusa.net, this
	// handler runs on prod even for dev/staging logins, proxied via
	// GetLoginForStaging's relay. If state is present and we're on prod,
	// this is a relay in progress: loop back into /login/staging (now
	// logged in) to complete the handoff to the dev/staging instance rather
	// than stopping at prod's own PostLoginURL. If state is present and
	// we're not on prod, this must be a true local-dev instance talking to
	// VATSIM directly (no relay to continue), so redirect straight there.
	if state := c.QueryParam("state"); state != "" && config.IsAllowedRedirect(state) {
		if config.IsProduction() {
			return c.Redirect(http.StatusFound, "/login/staging?redirect="+url.QueryEscape(state))
		}
		return c.Redirect(http.StatusFound, state)
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

	userModel := models.UserFromDatabase(*user, true)

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

func GetLoginForStaging(c *echo.Context) error {
	if !config.IsProduction() {
		return echo.NewHTTPError(http.StatusNotFound, "Not found")
	}
	redirect := validatedRedirect(c)
	if !auth.IsLoggedIn(c) {
		return c.Redirect(http.StatusFound, vatsim.ConnectFullURL(redirect))
	}
	cid := auth.GetUserCid(c)

	client := &http.Client{}
	tokenURL := fmt.Sprintf("%s/token/%d", config.StagingInternalURL(), cid)
	req, err := http.NewRequest("GET", tokenURL, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate staging token")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(auth.ActorTokenHeader, config.StagingActorToken())

	resp, err := client.Do(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate staging token")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate staging token")
	}

	data := make(map[string]string)
	err = json.Unmarshal(body, &data)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate staging token")
	}
	redirectUrl := fmt.Sprintf("%s/login/useToken/%s", config.StagingInternalURL(), data["token"])
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
