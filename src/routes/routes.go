package routes

import (
	"vatusa-cobalt/config"
	"vatusa-cobalt/dbconn"
	"vatusa-cobalt/endpoints"
	_middleware "vatusa-cobalt/middleware"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func SetupRoutes(e *echo.Echo) {
	queries := dbconn.Queries()

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

	user := e.Group("/user")
	user.GET("/:cid", handler.GetUser)

	login := e.Group("/login")
	login.GET("", handler.GetLogin)
	login.GET("/connect", handler.Connect)
	login.GET("/as/:cid", handler.LoginAs)
	login.GET("/whoami", handler.WhoAmI)
	login.GET("/logout", handler.GetLogout)
	login.GET("/staging", handler.GetLoginForStaging)
	login.GET("/useToken/:token", handler.LoginUseToken)

	my := e.Group("/my")
	my.GET("/session", handler.GetMySession)

}
