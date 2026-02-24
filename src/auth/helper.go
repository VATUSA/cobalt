package auth

import "github.com/labstack/echo/v5"

func GetUserCid(c *echo.Context) int {
	cid, ok := c.Get(ContextUserCID).(int)
	if !ok {
		return -1
	}
	return cid
}

func IsLoggedIn(c *echo.Context) bool {
	return GetUserCid(c) > 0
}

func GetActorId(c *echo.Context) int {
	token, ok := c.Get(ContextActorId).(int)
	if !ok {
		return -1
	}
	return token
}
