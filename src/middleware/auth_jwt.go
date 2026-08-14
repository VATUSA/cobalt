package middleware

import (
	"net/http"
	"vatusa-cobalt/auth"

	"github.com/labstack/echo/v5"
)

func CookieAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		cid := -1
		token, err := c.Cookie(auth.JWTCookieName)
		if err == nil {
			cid, err = auth.GetCIDFromToken(token.Value)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired session")
			}
		}
		c.Set(auth.ContextUserCID, cid)
		return next(c)
	}
}

func RequireLogin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		if !auth.IsLoggedIn(c) {
			return c.String(http.StatusUnauthorized, "login required")
		}
		return next(c)
	}
}
