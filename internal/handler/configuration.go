package handler

import (
	"bytes"
	"context"
	"fmt"
	htmlTemplate "html/template"
	"io"
	"net/http"
	"os"
	"strconv"

	//templ "github.com/a-h/templ"
	"github.com/crocoder-dev/intro-video/internal"
	"github.com/crocoder-dev/intro-video/internal/config"
	"github.com/crocoder-dev/intro-video/internal/data"
	"github.com/crocoder-dev/intro-video/internal/template"
	"github.com/crocoder-dev/intro-video/internal/template/shared"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"
	"github.com/oklog/ulid/v2"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

func Configuration(c echo.Context) error {
	defaultConfig := config.IntroVideoFormValues{
        Url:        "",
        Theme:      config.DefaultTheme,
        CtaEnabled:        false,
        BubbleEnabled:     false,
        CtaText:    "",
        BubbleText: "",
    }

    id := c.Param("ulid")

    configuration := defaultConfig

    if id != "" && id != "new" {
        err := godotenv.Load(".env")
        if err != nil {
            return err
        }

        dbUrl := os.Getenv("DATABASE_URL")
        authToken := os.Getenv("TURSO_AUTH_TOKEN")
        if dbUrl == "" || authToken == "" {
            return fmt.Errorf("DATABASE_URL and TURSO_AUTH_TOKEN must be set in .env file")
        }

        store := data.Store{DatabaseUrl: dbUrl + "?authToken=" + authToken, DriverName: "libsql"}

        byteId, err := ulid.Parse(id)
        if err != nil {
            fmt.Printf("Failed to parse ULID: %v. Using default configuration.\n", err)
        } else {
            loadedConfig, err := store.LoadConfig(byteId.Bytes())
            if err != nil {
                fmt.Printf("Failed to load configuration: %v. Using default configuration.\n", err)
            } else {
                configuration = config.IntroVideoFormValues{
					Theme: loadedConfig.Theme,
					CtaEnabled: loadedConfig.Cta.Enabled,
					BubbleEnabled: loadedConfig.Bubble.Enabled,
					BubbleText: loadedConfig.Bubble.TextContent,
					CtaText: loadedConfig.Cta.TextContent,
					Url: loadedConfig.VideoUrl,
				}
            }
        }
    }

    themeOptions := []template.ThemeOption{
        {Caption: "Default Theme", Value: config.DefaultTheme, Selected: true},
        {Caption: "Shadcn Theme - Light", Value: config.ShadcnThemeLight},
        {Caption: "Shadcn Theme - Dark", Value: config.ShadcnThemeDark},
        {Caption: "MaterialUi Theme - Light", Value: config.MaterialUiThemeLight},
        {Caption: "MaterialUi Theme - Dark", Value: config.MaterialUiThemeDark},
        {Caption: "Tailwind Theme - Dark", Value: config.TailwindThemeDark},
        {Caption: "Tailwind Theme - Light", Value: config.TailwindThemeLight},
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
    basePreviewJs := "<script>" + string(base) + "</script>"
    component := template.Configuration(
        themeOptions,
        basePreviewJs,
        configuration,
		id,
    )
	return component.Render(context.Background(), c.Response().Writer)

}

func initializeStore() (data.Store, error) {
	err := godotenv.Load(".env")
	if err != nil {
		return data.Store{}, err
	}
	dbUrl := os.Getenv("DATABASE_URL")
	authToken := os.Getenv("TURSO_AUTH_TOKEN")
	if dbUrl == "" || authToken == "" {
		return data.Store{}, fmt.Errorf("DATABASE_URL and TURSO_AUTH_TOKEN must be set in .env file")
	}
	return data.Store{DatabaseUrl: dbUrl + "?authToken=" + authToken, DriverName: "libsql"}, nil
}

func parseFormValues(c echo.Context) (data.NewConfiguration, error) {
	newVideo := data.NewVideo{Weight: 100, URL: c.FormValue(template.URL)}
	newTheme := config.Theme(c.FormValue(template.THEME))

	newBubbleEnabledStr := c.FormValue(template.BUBBLE_ENABLED)
	if newBubbleEnabledStr == "" {
		newBubbleEnabledStr = "false"
	}
	newBubbleEnabled, err := strconv.ParseBool(newBubbleEnabledStr)
	if err != nil {
		return data.NewConfiguration{}, err
	}

	newCtaEnabledStr := c.FormValue(template.CTA_ENABLED)
	if newCtaEnabledStr == "" {
		newCtaEnabledStr = "false"
	}

	newCtaEnabled, err := strconv.ParseBool(newCtaEnabledStr)
	if err != nil {
		return data.NewConfiguration{}, err
	}

	return data.NewConfiguration{
		VideoUrl: newVideo.URL,
		Theme: newTheme,
		Bubble: config.Bubble{
			Enabled:     newBubbleEnabled,
			TextContent: c.FormValue(template.BUBBLE_TEXT),
		},
		Cta: config.Cta{
			Enabled:     newCtaEnabled,
			TextContent: c.FormValue(template.CTA_TEXT),
		},
	}, nil
}

func CreateConfig(c echo.Context) error {
	store, err := initializeStore()
	if err != nil {
		return err
	}

	newConfiguration, err := parseFormValues(c)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	config, err := store.CreateConfiguration(newConfiguration)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	redirectURL := fmt.Sprintf("/v/%v", ulid.ULID(config.Id).String())
	c.Response().Header().Set("HX-Redirect", redirectURL)
	return c.NoContent(http.StatusOK)
}

func UpdateConfig(c echo.Context) error {
	id := c.Param("ulid")
	if id == "" {
		return c.String(http.StatusBadRequest, "Missing configuration ID")
	}

	store, err := initializeStore()
	if err != nil {
		return err
	}

	updatedConfiguration, err := parseFormValues(c)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	configID, err := ulid.Parse(id)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid configuration ID")
	}

	_, err = store.UpdateConfiguration(configID.Bytes(), updatedConfiguration)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	redirectURL := fmt.Sprintf("/v/%v", id)
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

	if bubbleEnabledRaw == "" || bubbleEnabledRaw == "off" {
		bubbleEnabled = false
	} else if bubbleEnabledRaw == "on" || bubbleEnabledRaw == "true" {
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
	if ctaEnabledRaw == "on" || ctaEnabledRaw == "true" {
		ctaEnabled = true
	} else if ctaEnabledRaw == "off" ||  ctaEnabledRaw == "" {
		ctaEnabled = false
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

	exportScript, err := internal.Script{}.Process(processableFileProps, internal.ProcessableFileOpts{Export: true, Minify: true})
	exportScript = "<script>" + exportScript + "</script>"

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
	shared.SuccessToast("errorko").Render(context.Background(), c.Response().Writer)

	component := template.IntroVideoPreview(js, css, previewScript, previewStyle, exportScript)
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
