package endpoints

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
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

func GetLogin(c *echo.Context) error {
	if config.IsStaging() {
		return c.Redirect(http.StatusFound, "https://cobalt.vatusa.net/login/staging")
	}
	return c.Redirect(http.StatusFound, vatsim.ConnectFullURL())
}

func GetLogout(c *echo.Context) error {
	c.SetCookie(&http.Cookie{
		Name:   auth.JWTCookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
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
	c.SetCookie(&http.Cookie{
		Name:     auth.JWTCookieName,
		Value:    jwt,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(time.Hour * 24 * 7),
	})

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

func GetUserDetailsFromToken(c *echo.Context) error {
	if !AssertGlobal(c, acl.ObjectLegacyLoginToken, acl.ActionRead) {
		return nil
	}
	tokenString := c.Param("token")
	cid, err := auth.GetCIDFromToken(tokenString)
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
	c.SetCookie(&http.Cookie{
		Name:  auth.JWTCookieName,
		Value: token,
		Path:  "/",
	})
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
	if !auth.IsLoggedIn(c) {
		return c.Redirect(http.StatusFound, vatsim.ConnectFullURL())
	}
	cid := auth.GetUserCid(c)

	client := &http.Client{}
	url := fmt.Sprintf("%s/token/%d", config.StagingInternalURL(), cid)
	req, err := http.NewRequest("GET", url, nil)
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
	return c.Redirect(http.StatusFound, redirectUrl)
}

func LoginUseToken(c *echo.Context) error {
	if !config.IsStaging() {
		return echo.NewHTTPError(http.StatusNotFound, "Not found")
	}
	token := c.Param("token")

	c.SetCookie(&http.Cookie{
		Name:  auth.JWTCookieName,
		Value: token,
		Path:  "/",
	})
	return c.Redirect(http.StatusFound, config.PostLoginURL())
}
