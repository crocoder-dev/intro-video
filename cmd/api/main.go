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

	e.GET("/v/:ulid", handler.Configuration)
	e.GET("/v/new", handler.Configuration)
	e.POST("/v/new", handler.IntroVideoCode)
	e.GET("/sign-in", handler.SignIn)

	e.File("/", "internal/template/demo.html")

	e.Static("/", "public")

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	e.Logger.Fatal(e.Start(":" + port))
}
