package endpoints

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
	"vatusa-cobalt/auth"
	"vatusa-cobalt/config"

	"github.com/labstack/echo/v5"
)

func (h EndpointHandler) GetLogin(c *echo.Context) error {
	return c.Redirect(http.StatusFound, auth.ConnectFullURL())
}

func (h EndpointHandler) GetLogout(c *echo.Context) error {
	c.SetCookie(&http.Cookie{
		Name:   auth.JWT_COOKIE_NAME,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
		Domain: config.CookieDomain(),
	})
	return c.Redirect(http.StatusFound, config.PostLoginURL())
}

func (h EndpointHandler) Connect(c *echo.Context) error {
	code := c.QueryParam("code")
	token, err := auth.FetchToken(code)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "error fetching connect access token")
	}
	userData, err := auth.FetchUserData(token)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "error fetching connect user data")
	}
	cid, err := strconv.Atoi(userData.CID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "error extracting cid")
	}

	// TODO: use userData to update the user record

	jwt, err := auth.CreateTokenForCID(cid)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create token")
	}
	c.SetCookie(&http.Cookie{
		Name:     auth.JWT_COOKIE_NAME,
		Value:    jwt,
		Path:     "/",
		Domain:   config.CookieDomain(),
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(time.Hour * 24 * 7),
	})

	return c.Redirect(http.StatusFound, config.PostLoginURL())
}

func (h EndpointHandler) GetGenerateUserToken(c *echo.Context) error {
	actor := auth.GetTokenActor(c)
	if !actor.HasGlobalGrant(auth.GrantGenerateUserAuthToken) {
		return echo.NewHTTPError(http.StatusForbidden, "not authorized")
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
		Name:  auth.JWT_COOKIE_NAME,
		Value: token,
		Path:  "/",
	})
	return c.Redirect(http.StatusFound, config.PostLoginURL())
}

func (h EndpointHandler) WhoAmI(c *echo.Context) error {
	cid := auth.GetUserCid(c)

	return c.String(http.StatusOK, fmt.Sprintf("%d", cid))
}
