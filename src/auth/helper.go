package auth

import "github.com/labstack/echo/v5"

func GetUserCid(c *echo.Context) int {
	cid, ok := c.Get(CONTEXT_USER_CID).(int)
	if !ok {
		return -1
	}
	return cid
}

func IsLoggedIn(c *echo.Context) bool {
	return GetUserCid(c) > 0
}

func GetTokenActor(c *echo.Context) *TokenActor {
	tokenActor, ok := c.Get(ContextTokenActor).(TokenActor)
	if !ok {
		return nil
	}
	return &tokenActor
}
