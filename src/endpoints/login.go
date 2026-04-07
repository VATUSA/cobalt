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
	"vatusa-cobalt/vatsim"

	"github.com/labstack/echo/v5"
)

func (h EndpointHandler) GetLogin(c *echo.Context) error {
	if config.IsStaging() {
		return c.Redirect(http.StatusFound, "https://cobalt.vatusa.net/login/staging")
	}
	return c.Redirect(http.StatusFound, vatsim.ConnectFullURL())
}

func (h EndpointHandler) GetLogout(c *echo.Context) error {
	c.SetCookie(&http.Cookie{
		Name:   auth.JWTCookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
		Domain: config.CookieDomain(),
	})
	return c.Redirect(http.StatusFound, config.PostLoginURL())
}

func (h EndpointHandler) Connect(c *echo.Context) error {
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

	if config.IsProduction() || config.IsStaging() {
		job := background.NewJob("vatsim_sync", fmt.Sprintf("%d", cid))
		err = job.Run()
		if err != nil {
			log.Printf("error syncing user cid %d: %v", cid, err)
		}
	}

	jwt, err := auth.CreateTokenForCID(cid)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create token")
	}
	c.SetCookie(&http.Cookie{
		Name:     auth.JWTCookieName,
		Value:    jwt,
		Path:     "/",
		Domain:   config.CookieDomain(),
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(time.Hour * 24 * 7),
	})

	return c.Redirect(http.StatusFound, config.PostLoginURL())
}

func (h EndpointHandler) GetGenerateUserToken(c *echo.Context) error {
	if !h.AssertGlobal(c, acl.ObjectLegacyLoginToken, acl.ActionWrite) {
		return nil
	}
	cid, err := strconv.Atoi(c.Param("cid"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid cid")
	}
	token, err := auth.CreateTokenForCID(cid)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create token")
	}
	data := make(map[string]string)
	data["token"] = token
	return c.JSON(http.StatusOK, data)
}

func (h EndpointHandler) LoginAs(c *echo.Context) error {
	if !config.IsDevelopment() {
		return echo.NewHTTPError(http.StatusNotFound, "Not found")
	}
	cid, err := strconv.Atoi(c.Param("cid"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad Request")
	}

	token, err := auth.CreateTokenForCID(cid)
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

func (h EndpointHandler) WhoAmI(c *echo.Context) error {
	cid := auth.GetUserCid(c)

	return c.String(http.StatusOK, fmt.Sprintf("%d", cid))
}

func (h EndpointHandler) GetLoginForStaging(c *echo.Context) error {
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
	redirectUrl := fmt.Sprintf("%s/login/useToken/%d", config.StagingInternalURL(), data["token"])
	return c.Redirect(http.StatusFound, redirectUrl)
}

func (h EndpointHandler) LoginUseToken(c *echo.Context) error {
	if !config.IsStaging() {
		return echo.NewHTTPError(http.StatusNotFound, "Not found")
	}
	token := c.Param("token")

	c.SetCookie(&http.Cookie{
		Name:  auth.JWTCookieName,
		Value: token,
		Path:  "/",
	})
	return c.JSON(http.StatusOK, "success")
}
