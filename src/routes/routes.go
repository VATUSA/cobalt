package routes

import (
	"database/sql"
	"log"
	"vatusa-cobalt/config"
	"vatusa-cobalt/db"
	"vatusa-cobalt/endpoints"
	_middleware "vatusa-cobalt/middleware"

	"github.com/labstack/echo/v5"
)

func SetupRoutes(e *echo.Echo) {
	database, err := sql.Open("mysql", config.ConnectionString())
	if err != nil {
		log.Fatal(err)
	}
	queries := db.New(database)

	handler := endpoints.NewEndpointHandler(queries)
	actorAuth := _middleware.NewActorAuth(queries)

	api := e.Group("/api")
	api.Use(actorAuth.Middleware)
	api.GET("/token/:cid", handler.GetGenerateUserToken)
	api.POST("/roles/legacy_sync", handler.LegacySyncRoles)

	login := e.Group("/login")
	login.Use(_middleware.CookieAuth)
	login.GET("", handler.GetLogin)
	login.GET("/connect", handler.Connect)
	login.GET("/as/:cid", handler.LoginAs)
	login.GET("/whoami", handler.WhoAmI)

	web := e.Group("/web")
	web.Use(_middleware.CookieAuth)
	web.GET("/news/:count", handler.GetLastPosts)
	web.POST("/news/new", handler.CreatePost)
}
