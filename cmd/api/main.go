package main

import (
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/crocoder-dev/intro-video/internal/data"
	"github.com/crocoder-dev/intro-video/internal/handler"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()
	err := applyMigrations()
	if err != nil {
		e.Logger.Fatal(err.Error())
		panic(err)
	}

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/demo/script.js", handler.Script)

	e.GET("/demo/style.css", handler.Stylesheet)

	e.GET("/v/:ulid", handler.Configuration)
	e.GET("/v/new", handler.Configuration)
	e.POST("/v/new", handler.IntroVideoCode)
	e.POST("/v/config", handler.ConfigSave)

	e.File("/", "internal/template/demo.html")

	e.Static("/", "public")

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	e.Logger.Fatal(e.Start(":" + port))
}

func applyMigrations() error {
	migrationsPath := filepath.Join("db", "migrations")

	var schemaFiles []string
	err := filepath.WalkDir(migrationsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, ".down.sql") {
			// schemaFiles = append(schemaFiles, path)
		}

		if strings.HasSuffix(path, ".up.sql") {
			// schemaFiles = append(schemaFiles, path)
		}

		return nil
	})

	if err != nil {
		return err
	}

	sort.Strings(schemaFiles)
	erra := godotenv.Load(".env")
	if erra != nil {
		return fmt.Errorf("env not loaded!")
	}
	dbUrl := os.Getenv("DATABASE_URL")
	authToken := os.Getenv("TURSO_AUTH_TOKEN")

	if dbUrl == "" || authToken == "" {
		return fmt.Errorf("DATABASE_URL and TURSO_AUTH_TOKEN must be set in .env file")
	}
	store := data.Store{DatabaseUrl: dbUrl + "?authToken=" + authToken, DriverName: "libsql"}

	db, err := sql.Open(store.DriverName, store.DatabaseUrl)
	if err != nil {
		return err
	}

	for _, schemaFile := range schemaFiles {
		schema, err := os.ReadFile(schemaFile)

		if err != nil {
			return err
		}

		_, err = db.Exec(string(schema))

		if err != nil {
			return err
		}
	}
	return nil
}
