package handler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/crocoder-dev/intro-video/internal"
	"github.com/crocoder-dev/intro-video/internal/config"
	"github.com/crocoder-dev/intro-video/internal/template"
	"github.com/labstack/echo/v4"
)

func Configuration(c echo.Context) error {
	_ = c.Param("ulid")

	themeOptions := []template.ThemeOption{
		{Caption: "Default Theme", Value: config.DefaultTheme, Selected: true},
	}

	file, err := os.Open("internal/template/script/base.js")
	if err != nil {
		return err
	}
	defer file.Close()

	base, err := io.ReadAll(file)

	var result bytes.Buffer
	result.Write(base)

	basePreviewJs := "<script>" + result.String() + "</script>"

	component := template.Configuration(themeOptions, basePreviewJs)
	return component.Render(context.Background(), c.Response().Writer)
}

func IntroVideoCode(c echo.Context) error {
	fmt.Println(
		"url", c.FormValue(template.URL), "\n",
		"bubbleEnabled", c.FormValue(template.BUBBLE_ENABLED), "\n",
		"bubbleText", c.FormValue(template.BUBBLE_TEXT), "\n",
		"theme", c.FormValue(template.THEME), "\n",
		"ctaEnabled", c.FormValue(template.CTA_ENABLED), "\n",
		"ctaText", c.FormValue(template.CTA_TEXT),
	)

	url := c.FormValue(template.URL)

	theme, err := config.NewTheme(c.FormValue(template.THEME))
	if err != nil {
		return err
	}

	bubbleEnabledRaw := c.FormValue(template.BUBBLE_ENABLED)

	var bubbleEnabled bool

	if bubbleEnabledRaw == "" {
		bubbleEnabled = false
	} else if bubbleEnabledRaw == "true" {
		bubbleEnabled = true
	}

	var bubbleTextContent string

	if bubbleEnabled {
		bubbleTextContent = c.FormValue(template.BUBBLE_TEXT)
	}

	var ctaEnabled bool

	ctaEnabledRaw := c.FormValue(template.CTA_ENABLED)

	if ctaEnabledRaw == "" {
		ctaEnabled = false
	} else if ctaEnabledRaw == "true" {
		ctaEnabled = true
	}

	var ctaTextContent string

	if ctaEnabled {
		ctaTextContent = c.FormValue(template.CTA_TEXT)
	}

	processableFileProps := internal.ProcessableFileProps{
		URL: url,
		Theme: theme,
		Bubble: config.Bubble{
			Enabled:     bubbleEnabled,
			TextContent: bubbleTextContent,
		},
		Cta: config.Cta{
			Enabled:     ctaEnabled,
			TextContent: ctaTextContent,
		},
	}

	previewScript, err := internal.Script{}.Process(processableFileProps, internal.ProcessableFileOpts{Preview: true})
	previewScript = "<script>" + previewScript + "</script>"

	previewStyle, err := internal.Stylesheet{}.Process(processableFileProps, internal.ProcessableFileOpts{Preview: true})
	previewStyle = "<style>" + previewStyle + "</style>"

	js, err := internal.Script{}.Process(processableFileProps, internal.ProcessableFileOpts{Minify: true})
	if err != nil {
		return err
	}

	css, err := internal.Stylesheet{}.Process(processableFileProps, internal.ProcessableFileOpts{Minify: true})
	if err != nil {
		return err
	}

	component := template.IntroVideoPreview(js, css, previewScript, previewStyle)
	return component.Render(context.Background(), c.Response().Writer)
}
