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
	e.Use(actorAuth.Middleware)
	e.Use(_middleware.CookieAuth)

	e.GET("/token/:cid", handler.GetGenerateUserToken)

	news := e.Group("/news")
	news.GET("/latest/:count", handler.GetLastPosts)
	news.GET("/page/:page", handler.GetNewsPage)
	news.POST("/new", handler.CreatePost)
	news.GET("/post/:id", handler.GetPost)
	news.POST("/post/:id", handler.UpdatePost)
	news.DELETE("/post/:id", handler.DeletePost)

	roles := e.Group("/roles")
	roles.POST("/legacy_sync", handler.LegacySyncRoles)
	roles.POST("/legacy_sync/bulk", handler.LegacySyncRolesBulk)

	event := e.Group("/event")
	event.GET("/upcoming/:count", handler.GetUpcomingEvents)
	event.GET("/page/:page", handler.GetEventsPage)
	event.GET("/:id", handler.GetEventById)
	event.POST("/create", handler.CreateEvent)
	event.POST("/:id", handler.UpdateEvent)
	event.DELETE("/:id", handler.DeleteEvent)

	api := e.Group("/api")
	api.GET("/token/:cid", handler.GetGenerateUserToken)
	api.POST("/roles/legacy_sync", handler.LegacySyncRoles)
	api.GET("/news/:id", handler.GetPost)
	api.GET("/news/latest/:count", handler.GetLastPosts)
	api.GET("/news/page/:page", handler.GetNewsPage)

	login := e.Group("/login")
	login.GET("", handler.GetLogin)
	login.GET("/connect", handler.Connect)
	login.GET("/as/:cid", handler.LoginAs)
	login.GET("/whoami", handler.WhoAmI)

}
