package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"

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

func GetConfiguration(c echo.Context) error {
	defaultConfig := config.IntroVideoFormValues{
		Url:           "",
		Theme:         config.DefaultTheme,
		CtaEnabled:    false,
		BubbleEnabled: false,
		CtaText:       "",
		BubbleText:    "",
	}

	id := c.Param("ulid")

	configuration := defaultConfig

	if id != "" && id != "new" {
		err := godotenv.Load(".env")
		if err != nil {
			return err
		}

		dbUrl := os.Getenv("DATABASE_URL")
		if dbUrl == "" {
			return fmt.Errorf("DATABASE_URL must be set in .env file")
		}

		store := data.Store{DatabaseUrl: dbUrl, DriverName: "libsql"}

		byteId, err := ulid.Parse(id)
		if err != nil {
			return shared.ErrorToast("Failed to parse ULID: %v. Using default configuration.").Render(context.Background(), c.Response().Writer)
		} else {
			loadedConfig, err := store.LoadConfiguration(byteId.Bytes())
			if err != nil {
				return shared.ErrorToast("Failed to load configuration: %v. Using default configuration.").Render(context.Background(), c.Response().Writer)
			} else {
				configuration = config.IntroVideoFormValues{
					Theme:         loadedConfig.Theme,
					CtaEnabled:    loadedConfig.Cta.Enabled,
					BubbleEnabled: loadedConfig.Bubble.Enabled,
					BubbleText:    loadedConfig.Bubble.TextContent,
					CtaText:       loadedConfig.Cta.TextContent,
					Url:           loadedConfig.VideoUrl,
				}
			}
		}
	}

	themeOptions := []template.ThemeOption{
		{Caption: "Default Theme", Value: config.DefaultTheme},
		{Caption: "Shadcn Theme - Light", Value: config.ShadcnThemeLight},
		{Caption: "Shadcn Theme - Dark", Value: config.ShadcnThemeDark},
		{Caption: "MaterialUi Theme - Light", Value: config.MaterialUiThemeLight},
		{Caption: "MaterialUi Theme - Dark", Value: config.MaterialUiThemeDark},
		{Caption: "Tailwind Theme - Dark", Value: config.TailwindThemeDark},
		{Caption: "Tailwind Theme - Light", Value: config.TailwindThemeLight},
		{Caption: "CroCoder Theme", Value: config.Crocoder, Selected: true},
		{Caption: "None", Value: config.NoneTheme},
	}

	file, err := os.Open("internal/template/script/base.js")
	if err != nil {
		return shared.ErrorToast("Something went wrong!").Render(context.Background(), c.Response().Writer)
	}
	defer file.Close()
	base, err := io.ReadAll(file)
	if err != nil {
		return shared.ErrorToast("Failed to read the base script file.").Render(context.Background(), c.Response().Writer)
	}
	basePreviewJS := "<script>" + string(base) + "</script>"
	videoFormProps := template.VideoFormProps{
		ThemeOptions:  themeOptions,
		BasePreviewJS: basePreviewJS,
		FormValues:    configuration,
		Ulid:          id,
	}
	videoPreviewProps := template.VideoPreviewProps{}

	component := template.Configuration(
		videoFormProps,
		videoPreviewProps,
	)
	return component.Render(context.Background(), c.Response().Writer)
}

func initializeStore() (data.Store, error) {
	err := godotenv.Load(".env")
	if err != nil {
		return data.Store{}, err
	}
	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		return data.Store{}, fmt.Errorf("DATABASE_URL must be set in .env file")
	}
	return data.Store{DatabaseUrl: dbUrl, DriverName: "libsql"}, nil
}

func parseFormValues(c echo.Context) (data.NewConfiguration, error) {
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
		VideoUrl: c.FormValue(template.URL),
		Theme:    newTheme,
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

func validateConfiguration(config data.NewConfiguration) error {
	if err := validateURL(config.VideoUrl); err != nil {
		return err
	}

	return nil
}

func validateURL(videoUrl string) error {
	if videoUrl == "" {
		return fmt.Errorf("Video url is empty!")
	}

	parsedUrl, err := url.ParseRequestURI(videoUrl)
	if err != nil || parsedUrl.Scheme == "" || parsedUrl.Host == "" {
		return fmt.Errorf("Video url is invalid!")
	}

	return nil
}

func CreateConfig(c echo.Context) error {
	store, err := initializeStore()
	if err != nil {
		return err
	}

	newConfiguration, err := parseFormValues(c)
	if err != nil {
		return shared.ErrorToast("Something went wrong!").Render(context.Background(), c.Response().Writer)
	}

	err = validateConfiguration(newConfiguration)
	if err != nil {
		return shared.ErrorToast(err.Error()).Render(context.Background(), c.Response().Writer)
	}

	config, err := store.CreateConfiguration(newConfiguration)
	if err != nil {

		return shared.ErrorToast(err.Error()).Render(context.Background(), c.Response().Writer)
	}

	redirectURL := fmt.Sprintf("/v/%v", ulid.ULID(config.Id).String())
	c.Response().Header().Set("HX-Redirect", redirectURL)
	return c.NoContent(http.StatusOK)
}

func UpdateConfig(c echo.Context) error {
	id := c.Param("ulid")
	if id == "" {
		return shared.ErrorToast("Missing configuration ID!").Render(context.Background(), c.Response().Writer)
	}

	store, err := initializeStore()
	if err != nil {
		return err
	}

	updatedConfiguration, err := parseFormValues(c)
	if err != nil {
		return shared.ErrorToast(err.Error()).Render(context.Background(), c.Response().Writer)
	}

	configID, err := ulid.Parse(id)
	if err != nil {
		return shared.ErrorToast("Invalid configuration ID").Render(context.Background(), c.Response().Writer)
	}

	_, err = store.UpdateConfiguration(configID.Bytes(), updatedConfiguration)
	if err != nil {
		return shared.ErrorToast(err.Error()).Render(context.Background(), c.Response().Writer)
	}

	redirectURL := fmt.Sprintf("/v/%v", id)
	c.Response().Header().Set("HX-Redirect", redirectURL)
	return c.NoContent(http.StatusOK)
}

func createVideoPreviewProps(videoUrl string, theme config.Theme, bubble config.Bubble, cta config.Cta) (template.VideoPreviewProps, error) {

	processableFileProps := internal.ProcessableFileProps{
		URL:    videoUrl,
		Theme:  theme,
		Bubble: bubble,
		Cta:    cta,
	}

	previewScript, err := internal.Script{}.Process(processableFileProps, internal.ProcessableFileOpts{Preview: true})
	previewScript = "<script>" + previewScript + "</script>"

	previewStyle, err := internal.Stylesheet{}.Process(processableFileProps, internal.ProcessableFileOpts{Preview: true})
	previewStyle = "<style>" + previewStyle + "</style>"

	js, err := internal.Script{}.Process(processableFileProps, internal.ProcessableFileOpts{Minify: true})
	if err != nil {
		return template.VideoPreviewProps{}, err
	}

	css, err := internal.Stylesheet{}.Process(processableFileProps, internal.ProcessableFileOpts{Minify: true})
	if err != nil {
		return template.VideoPreviewProps{}, err
	}
	return template.VideoPreviewProps{
		JS:            js,
		CSS:           css,
		PreviewScript: previewScript,
		PreviewStyle:  previewStyle,
	}, nil
}

func IntroVideoCode(c echo.Context) error {

	configuration, err := parseFormValues(c)

	if err != nil {
		return shared.ErrorToast("Something went wrong!").Render(context.Background(), c.Response().Writer)
	}

	err = validateConfiguration(configuration)
	if err != nil {
		return shared.ErrorToast(err.Error()).Render(context.Background(), c.Response().Writer)
	}

	videoPreviewProps, err := createVideoPreviewProps(configuration.VideoUrl, configuration.Theme, configuration.Bubble, configuration.Cta)
	if err != nil {
		return shared.ErrorToast(err.Error()).Render(context.Background(), c.Response().Writer)
	}

	component := template.IntroVideoPreview(videoPreviewProps)
	return component.Render(context.Background(), c.Response().Writer)
}
