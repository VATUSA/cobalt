package middleware

import (
	"context"
	"log"
	"net/http"
	"time"
	"vatusa-cobalt/auth"
	"vatusa-cobalt/db"

	"github.com/labstack/echo/v5"
)

type ActorAuth struct {
	Cache auth.TokenActorCache
}

func (a *ActorAuth) Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		tokenHeader := c.Request().Header.Get(auth.ActorTokenHeader)

		tokenActor, ok := a.Cache.Get(tokenHeader)
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, "Token required")
		}
		c.Set(auth.ContextTokenActor, *tokenActor)
		return next(c)
	}
}

func NewActorAuth(queries *db.Queries) *ActorAuth {
	a := ActorAuth{}

	go func() {
		ctx := context.Background()
		for {
			err := a.Cache.Load(ctx, queries)
			if err != nil {
				log.Printf("Error loading token cache: %v\n", err)
			}
			time.Sleep(time.Minute)
		}
	}()

	return &a
}
