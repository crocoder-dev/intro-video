package handler

import (
	"context"

	"github.com/crocoder-dev/intro-video/internal/template"
	"github.com/labstack/echo/v4"
)

func SignIn(c echo.Context) error {
	component := template.Modal()
	return component.Render(context.Background(), c.Response().Writer)
}
