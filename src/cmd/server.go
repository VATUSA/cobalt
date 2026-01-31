package main

import (
	"net/http"
	"vatusa-cobalt/config"
	"vatusa-cobalt/routes"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/joho/godotenv/autoload"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	m, err := migrate.New(
		"file://sql/migrations/",
		"mysql://"+config.ConnectionString())
	if err != nil {
		panic(err)
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		panic(err)
	}

	e := echo.New()
	e.Group("/api")
	e.Use(middleware.RequestLogger())

	e.GET("/", func(c *echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	routes.SetupRoutes(e)
	if err := e.Start("0.0.0.0:8080"); err != nil {
		e.Logger.Error("Failed to start server", "error", err)
	}
}
