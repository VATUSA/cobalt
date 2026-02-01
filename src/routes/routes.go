package routes

import (
	"database/sql"
	"log"
	"vatusa-cobalt/config"
	"vatusa-cobalt/db"
	"vatusa-cobalt/endpoints"
	_middleware "vatusa-cobalt/middleware"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func SetupRoutes(e *echo.Echo) {
	database, err := sql.Open("mysql", config.ConnectionString())
	if err != nil {
		log.Fatal(err)
	}
	queries := db.New(database)

	handler := endpoints.NewEndpointHandler(queries)
	actorAuth := _middleware.NewActorAuth(queries)

	e.Use(middleware.CORSWithConfig(config.CORSConfig()))

	api := e.Group("/api")
	api.Use(actorAuth.Middleware)
	api.GET("/token/:cid", handler.GetGenerateUserToken)
	api.POST("/roles/legacy_sync", handler.LegacySyncRoles)
	api.GET("/news/:id", handler.GetPost)
	api.GET("/news/latest/:count", handler.GetLastPosts)

	login := e.Group("/login")
	login.Use(_middleware.CookieAuth)
	login.GET("", handler.GetLogin)
	login.GET("/connect", handler.Connect)
	login.GET("/as/:cid", handler.LoginAs)
	login.GET("/whoami", handler.WhoAmI)

	web := e.Group("/web")
	web.Use(_middleware.CookieAuth)
	web.GET("/news/:count", handler.GetLastPosts)
	web.GET("/news/post/:id", handler.GetPost)

	webRequireLogin := web.Group("")
	webRequireLogin.Use(_middleware.RequireLogin)
	webRequireLogin.POST("/news/new", handler.CreatePost)

}
