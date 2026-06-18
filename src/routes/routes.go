package routes

import (
	"vatusa-cobalt/config"
	"vatusa-cobalt/endpoints"
	_middleware "vatusa-cobalt/middleware"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func SetupRoutes(e *echo.Echo) {
	actorAuth := _middleware.NewActorAuth()

	e.Use(middleware.CORSWithConfig(config.CORSConfig()))
	e.Use(actorAuth.Middleware)
	e.Use(_middleware.CookieAuth)

	e.GET("/token/:cid", endpoints.GetGenerateUserToken)
	e.POST("/tokenSession", endpoints.PostUserDetailsFromToken)

	news := e.Group("/news")
	news.GET("/latest/:count", endpoints.GetLastPosts)
	news.GET("/page/:page", endpoints.GetNewsPage)
	news.POST("/new", endpoints.CreatePost)
	news.GET("/post/:id", endpoints.GetPost)
	news.POST("/post/:id", endpoints.UpdatePost)
	news.DELETE("/post/:id", endpoints.DeletePost)

	roles := e.Group("/roles")
	roles.POST("/legacy_sync", endpoints.LegacySyncRoles)
	roles.POST("/legacy_sync/bulk", endpoints.LegacySyncRolesBulk)

	event := e.Group("/event")
	event.GET("/upcoming/:count", endpoints.GetUpcomingEvents)
	event.GET("/page/:page", endpoints.GetEventsPage)
	event.GET("/:id", endpoints.GetEventById)
	event.POST("/create", endpoints.CreateEvent)
	event.POST("/:id", endpoints.UpdateEvent)
	event.POST("/:id/review", endpoints.ReviewEvent)
	event.DELETE("/:id", endpoints.DeleteEvent)

	user := e.Group("/user/:cid")
	user.GET("", endpoints.GetUser)
	user.GET("/blockers", endpoints.GetUserBlockers)

	login := e.Group("/login")
	login.GET("", endpoints.GetLogin)
	login.GET("/connect", endpoints.Connect)
	login.GET("/as/:cid", endpoints.LoginAs)
	login.GET("/whoami", endpoints.WhoAmI)
	login.GET("/logout", endpoints.GetLogout)
	login.GET("/staging", endpoints.GetLoginForStaging)
	login.GET("/useToken/:token", endpoints.LoginUseToken)

	my := e.Group("/my")
	my.GET("/session", endpoints.GetMySession)
	my.POST("/transfer", endpoints.MySubmitTransferRequest)

	roster := e.Group("/roster/:facility")
	roster.GET("", endpoints.GetFacilityRoster)
	roster.GET("/transfer", endpoints.GetFacilityPendingTransfers)
	roster.POST("/transfer", endpoints.ActionFacilityPendingTransfer)

}
