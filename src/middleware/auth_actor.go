package middleware

import (
	"context"
	"log"
	"time"
	"vatusa-cobalt/auth"

	"github.com/labstack/echo/v5"
)

type ActorAuth struct {
	Cache auth.TokenActorCache
}

func (a *ActorAuth) Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		tokenHeader := c.Request().Header.Get(auth.ActorTokenHeader)

		tokenActor, ok := a.Cache.Get(tokenHeader)
		if ok {
			c.Set(auth.ContextActorId, tokenActor.ActorId)
		}
		return next(c)
	}
}

// NewActorAuth loads the actor token cache synchronously before returning,
// so a request handled the moment the server comes up still sees actor
// tokens instead of silently falling through as anonymous. Refresh after
// that runs on a ticker in the background.
func NewActorAuth() *ActorAuth {
	a := ActorAuth{}

	ctx := context.Background()
	if err := a.Cache.Load(ctx); err != nil {
		log.Printf("Error loading token cache: %v\n", err)
	}

	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			if err := a.Cache.Load(ctx); err != nil {
				log.Printf("Error loading token cache: %v\n", err)
			}
		}
	}()

	return &a
}
