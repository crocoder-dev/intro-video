package main

import (
	"os"

	"github.com/crocoder-dev/intro-video/internal/handler"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/demo/script.js", handler.Script)

	e.GET("/demo/style.css", handler.Stylesheet)

	e.GET("/v/:ulid", handler.GetConfiguration)
	e.GET("/v/new", handler.GetConfiguration)

	e.POST("/v/:ulid", handler.UpdateConfig)
	e.POST("/v/new", handler.IntroVideoCode)
	e.POST("/v/config", handler.CreateConfig)

	e.File("/", "internal/template/demo.html")

	e.Static("/", "public")

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	e.Logger.Fatal(e.Start(":" + port))
}
