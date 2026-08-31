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

	faq := e.Group("/faq")
	faq.GET("", endpoints.GetFaqs)
	faq.POST("/category", endpoints.CreateFaqCategory)
	faq.POST("/category/:id", endpoints.UpdateFaqCategory)
	faq.DELETE("/category/:id", endpoints.DeleteFaqCategory)
	faq.POST("/item", endpoints.CreateFaqItem)
	faq.POST("/item/:id", endpoints.UpdateFaqItem)
	faq.DELETE("/item/:id", endpoints.DeleteFaqItem)

	solo := e.Group("/solo")
	solo.GET("", endpoints.GetSoloCerts)
	solo.POST("", endpoints.CreateSoloCert)
	solo.POST("/:id", endpoints.UpdateSoloCert)
	solo.DELETE("/:id", endpoints.DeleteSoloCert)

	policy := e.Group("/policy")
	policy.GET("", endpoints.GetPolicies)
	policy.POST("/category", endpoints.CreatePolicyCategory)
	policy.POST("/category/:id", endpoints.UpdatePolicyCategory)
	policy.DELETE("/category/:id", endpoints.DeletePolicyCategory)
	policy.POST("/document", endpoints.CreatePolicyDocument)
	policy.POST("/document/:id", endpoints.UpdatePolicyDocument)
	policy.DELETE("/document/:id", endpoints.DeletePolicyDocument)

	roles := e.Group("/roles")
	roles.GET("/ace-team", endpoints.GetACETeam)
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

	user := e.Group("/user")
	user.GET("/search", endpoints.SearchUsers)
	user.GET("/:cid", endpoints.GetUser)
	user.GET("/:cid/blockers", endpoints.GetUserBlockers)
	user.POST("/:cid/role/:facility", endpoints.GrantRole)
	user.DELETE("/:cid/role/:facility/:role", endpoints.RevokeRole)

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
	my.GET("/roles/assignable", endpoints.GetMyAssignableRoles)
	my.POST("/transfer", endpoints.MySubmitTransferRequest)

	roster := e.Group("/roster/:facility")
	roster.GET("", endpoints.GetFacilityRoster)
	roster.GET("/staff", endpoints.GetFacilityStaff)
	roster.GET("/transfer", endpoints.GetFacilityPendingTransfers)
	roster.POST("/transfer", endpoints.ActionFacilityPendingTransfer)

	facility := e.Group("/facility/:facility")
	facility.GET("/v3/apikeys", endpoints.GetFacilityApiKeys)

	facilityTitles := e.Group("/facility/:facility/titles")
	facilityTitles.GET("", endpoints.GetFacilityTitles)
	facilityTitles.POST("", endpoints.CreateFacilityTitle)
	facilityTitles.DELETE("/:id", endpoints.DeleteFacilityTitle)

	userTitles := e.Group("/user/:cid/facility/:facility/titles")
	userTitles.GET("", endpoints.GetUserFacilityTitles)
	userTitles.POST("", endpoints.AssignUserFacilityTitles)
	userTitles.DELETE("/:id", endpoints.DeleteUserFacilityTitle)

}
