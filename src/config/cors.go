package config

import (
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func AllowedOrigins() []string {
	if IsDevelopment() {
		return []string{
			"http://127.0.0.1:8000",
		}
	}

	return []string{
		"https://vatusa.net",
	}
}

func CORSConfig() middleware.CORSConfig {
	return middleware.CORSConfig{
		AllowOrigins: AllowedOrigins(),
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderXCSRFToken,
			echo.HeaderXRequestedWith,
			echo.HeaderContentType,
			echo.HeaderCookie,
			echo.HeaderSetCookie,
		},
		AllowCredentials: true,
	}
}
