package handler

import (
	"bytes"
	"context"
	"fmt"
	htmlTemplate "html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/crocoder-dev/intro-video/internal"
	"github.com/crocoder-dev/intro-video/internal/config"
	"github.com/crocoder-dev/intro-video/internal/data"
	"github.com/crocoder-dev/intro-video/internal/template"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"
	"github.com/oklog/ulid/v2"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

func Configuration(c echo.Context) error {
	_ = c.Param("ulid")

	themeOptions := []template.ThemeOption{
		{Caption: "Default Theme", Value: config.DefaultTheme, Selected: true},
		{Caption: "Shadcn Theme - Light", Value: config.ShadcnThemeLight},
		{Caption: "Shadcn Theme - Dark", Value: config.ShadcnThemeDark},
		{Caption: "MaterialUi Theme - Light", Value: config.MaterialUiThemeLight},
		{Caption: "MaterialUi Theme - Dark", Value: config.MaterialUiThemeDark},
		{Caption: "None", Value: config.NoneTheme},
	}

	file, err := os.Open("internal/template/script/base.js")
	if err != nil {
		return err
	}
	defer file.Close()

	base, err := io.ReadAll(file)

	if err != nil {
		return generateMessage(c, "Failed to read the base script file.", http.StatusInternalServerError)
	}

	var result bytes.Buffer
	result.Write(base)

	basePreviewJs := "<script>" + result.String() + "</script>"

	component := template.Configuration(themeOptions, basePreviewJs)
	return component.Render(context.Background(), c.Response().Writer)
}

func ConfigSave(c echo.Context) error {
	err := godotenv.Load(".env")
	if err != nil {
		return err
	}
	dbUrl := os.Getenv("DATABASE_URL")
	authToken := os.Getenv("TURSO_AUTH_TOKEN")

	if dbUrl == "" || authToken == "" {
		log.Fatal("DATABASE_URL and TURSO_AUTH_TOKEN must be set in .env file")
	}
	store := data.Store{DatabaseUrl: dbUrl+"?authToken="+authToken, DriverName: "libsql"}

	newVideo := data.NewVideo{Weight: 100, URL: c.FormValue(template.URL)}

	newTheme := config.Theme(c.FormValue(template.THEME))
	newBubbleEnabled, err := strconv.ParseBool(c.FormValue(template.BUBBLE_ENABLED))
	newCtaEnabled, err := strconv.ParseBool(c.FormValue(template.CTA_ENABLED))

	newConfiguration := data.NewConfiguration{
		Theme: newTheme,
		Bubble: config.Bubble{
			Enabled: newBubbleEnabled,
			TextContent: c.FormValue(template.BUBBLE_TEXT),
		},
		Cta: config.Cta{
			Enabled: newCtaEnabled,
			TextContent: c.FormValue(template.CTA_TEXT),
		},
	}

	instance, err := store.CreateInstance(newVideo, newConfiguration)
	if err != nil {
		return c.String(200, err.Error())
	}
	redirectURL := fmt.Sprintf("/v/%v", ulid.ULID(instance.ExternalId).String())
	c.Response().Header().Set("HX-Redirect", redirectURL)
	return c.NoContent(http.StatusOK)

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

	if url == "" {
		url = template.DEFAULT_URL
	}

	theme, err := config.NewTheme(c.FormValue(template.THEME))
	if err != nil {
		fmt.Println(err)
		return generateMessage(c, "There was an issue generating the theme. Please check the theme value and try again.", http.StatusInternalServerError)
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
		if c.FormValue(template.BUBBLE_TEXT) != "" {
			bubbleTextContent = c.FormValue(template.BUBBLE_TEXT)
		} else {
			bubbleTextContent = template.DEFAULT_BUBBLE_TEXT
		}
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
		if c.FormValue(template.CTA_TEXT) != "" {
			ctaTextContent = c.FormValue(template.CTA_TEXT)
		} else {
			ctaTextContent = template.DEFAULT_CTA_TEXT
		}
	}

	processableFileProps := internal.ProcessableFileProps{
		URL:   url,
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
		fmt.Println(err)
		return generateMessage(c, "An error occurred while generating the script. Please try again later.", http.StatusInternalServerError)
	}

	css, err := internal.Stylesheet{}.Process(processableFileProps, internal.ProcessableFileOpts{Minify: true})
	if err != nil {
		fmt.Println(err)
		return generateMessage(c, "An error occurred while generating the stylesheet. Please try again later.", http.StatusInternalServerError)
	}

	component := template.IntroVideoPreview(js, css, previewScript, previewStyle)
	return component.Render(context.Background(), c.Response().Writer)
}

const toastMessageTemplate = `
	<div class="pointer-events-auto w-full min-w-52 overflow-hidden rounded-lg bg-white shadow-lg ring-1 ring-black ring-opacity-5">
      	<div class="p-4">
        	<div class="flex items-start">
          		<div class="flex-shrink-0">
		  			<svg class="h-6 w-6 text-green-400" viewBox="0 0 20 20" fill="none" xmlns="http://www.w3.org/2000/svg">
		  				<circle cx="10" cy="10" r="9" stroke="#ef4444" stroke-width="2" fill="#ef4444"></circle>
		  				<path d="M7 7L13 13M13 7L7 13" stroke="#fff" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"></path>
	  				</svg>
				</div>
				<div class="ml-3 flex-1 pt-0.5">
					<p class="text-sm font-medium text-gray-900">Server error!</p>
					<p class="mt-1 text-sm text-gray-500">{{.Message}}</p>
          		</div>
        	</div>
      	</div>
    </div>
`

var tmpl = htmlTemplate.Must(htmlTemplate.New("toastMessage").Parse(toastMessageTemplate))

type ToastData struct {
	Message string
}

func generateMessageHtml(message string) (string, error) {
	data := ToastData{
		Message: message,
	}

	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, data); err != nil {
		return "", err
	}
	return tpl.String(), nil
}

func generateMessage(c echo.Context, message string, status int) error {
	html, err := generateMessageHtml(message)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Internal Server Error")
	}
	return c.HTML(status, html)
}
