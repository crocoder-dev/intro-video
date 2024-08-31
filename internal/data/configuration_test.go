package data_test

import (
	"database/sql"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/crocoder-dev/intro-video/internal/config"
	"github.com/crocoder-dev/intro-video/internal/data"
	_ "github.com/mattn/go-sqlite3"
	"github.com/oklog/ulid/v2"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

func getMigrationSchemas() ([]string, error) {
	migrationsPath := filepath.Join("..", "..", "db", "migrations")

	var schemaFiles []string

	err := filepath.WalkDir(migrationsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, ".up.sql") {
			schemaFiles = append(schemaFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	sort.Strings(schemaFiles)

	return schemaFiles, nil
}

func setupTestDB(t *testing.T) (*sql.DB, string) {
	file, err := os.CreateTemp("", "test*.db")
	if err != nil {
		t.Fatalf("failed to create database file: %v", err)
	}

	db, err := sql.Open("sqlite3", file.Name())
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	return db, file.Name()
}

func applySchemas(t *testing.T, db *sql.DB) {
	schemaFiles, err := getMigrationSchemas()
	if err != nil {
		t.Fatalf("failed to read schema files: %v", err)
	}

	for _, schemaFile := range schemaFiles {
		schema, err := os.ReadFile(schemaFile)
		if err != nil {
			t.Fatalf("failed to read schema file %s: %v", schemaFile, err)
		}

		_, err = db.Exec(string(schema))
		if err != nil {
			t.Fatalf("failed to execute schema %s: %v", schemaFile, err)
		}
	}
}

func insertTestData(t *testing.T, db *sql.DB, binUlid []byte) {
	_, err := db.Exec(`
		INSERT INTO configurations (id, bubble_enabled, bubble_text_content, cta_enabled, cta_text_content, theme, video_url)
		VALUES (?, 1, "bubble text", 1, "cta text", "default", "video url");
	`, binUlid)

	if err != nil {
		t.Fatalf("failed to insert test data: %v", err)
	}
}

func equalConfigs(expected, actual data.Configuration) bool {
	if string(expected.Id) != string(actual.Id) ||
		expected.Theme != actual.Theme ||
		expected.Bubble.Enabled != actual.Bubble.Enabled ||
		expected.Bubble.TextContent != actual.Bubble.TextContent ||
		expected.Cta.Enabled != actual.Cta.Enabled ||
		expected.Cta.TextContent != actual.Cta.TextContent ||
		expected.VideoUrl != actual.VideoUrl {
		return false
	}
	return true
}

func TestCreateConfiguration(t *testing.T) {
	db, dbName := setupTestDB(t)
	defer os.Remove(dbName)
	defer db.Close()

	applySchemas(t, db)

	store := data.Store{DatabaseUrl: dbName, DriverName: "sqlite3"}

	newConfiguration := data.NewConfiguration{
		Theme: config.DefaultTheme,
		Bubble: config.Bubble{
			Enabled:     true,
			TextContent: "bubble text",
		},
		Cta: config.Cta{
			Enabled:     true,
			TextContent: "cta text",
		},
		VideoUrl: "url",
	}

	configuration, err := store.CreateConfiguration(newConfiguration)
	if err != nil {
		t.Fatalf("failed to create instance: %v", err)
	}

	expected := data.Configuration{
		Id:       configuration.Id,
		Theme:    configuration.Theme,
		Cta:      configuration.Cta,
		Bubble:   configuration.Bubble,
		VideoUrl: configuration.VideoUrl,
	}

	if !equalConfigs(expected, configuration) {
		t.Fatalf("Expected configuration %+v, got %+v", expected, configuration)
	}
}

func TestLoadConfiguration(t *testing.T) {
	db, dbName := setupTestDB(t)
	defer os.Remove(dbName)
	defer db.Close()

	applySchemas(t, db)

	newUlid := ulid.Make()
	binUlid := newUlid.Bytes()

	insertTestData(t, db, binUlid)

	store := data.Store{DatabaseUrl: dbName, DriverName: "sqlite3"}

	configuration, err := store.LoadConfiguration(binUlid)
	if err != nil {
		t.Fatalf("failed to load instance: %v", err)
	}

	expected := data.Configuration{
		Id:    binUlid,
		Theme: config.DefaultTheme,
		Bubble: config.Bubble{
			Enabled:     true,
			TextContent: "bubble text",
		},
		Cta: config.Cta{
			Enabled:     true,
			TextContent: "cta text",
		},
		VideoUrl: "video url",
	}

	if !equalConfigs(expected, configuration) {
		t.Fatalf("Expected configuration %+v, got %+v", expected, configuration)
	}

	expectedUlid := ulid.Make()
	err = expectedUlid.UnmarshalBinary(configuration.Id)
	if err != nil {
		t.Fatalf("failed to unmarshal ulid: %v", err)
	}

	if newUlid.Compare(expectedUlid) != 0 {
		t.Fatalf("Expected ulid %s, got %s", expectedUlid, newUlid)
	}
}

func TestUpdateConfiguration(t *testing.T) {
	db, dbName := setupTestDB(t)
	defer os.Remove(dbName)
	defer db.Close()

	applySchemas(t, db)

	store := data.Store{DatabaseUrl: dbName, DriverName: "sqlite3"}

	newConfiguration := data.NewConfiguration{
		Theme: config.DefaultTheme,
		Bubble: config.Bubble{
			Enabled:     true,
			TextContent: "bubble text",
		},
		Cta: config.Cta{
			Enabled:     true,
			TextContent: "cta text",
		},
		VideoUrl: "url",
	}

	configuration, err := store.CreateConfiguration(newConfiguration)
	if err != nil {
		t.Fatalf("failed to create instance: %v", err)
	}

	updatedConfiguration := data.NewConfiguration{
		Theme: config.ShadcnThemeDark,
		Bubble: config.Bubble{
			Enabled:     false,
			TextContent: "updated bubble text",
		},
		Cta: config.Cta{
			Enabled:     false,
			TextContent: "updated cta text",
		},
		VideoUrl: "updated url",
	}

	expected := data.Configuration{
		Id:    configuration.Id,
		Theme: config.ShadcnThemeDark,
		Bubble: config.Bubble{
			Enabled:     false,
			TextContent: "updated bubble text",
		},
		Cta: config.Cta{
			Enabled:     false,
			TextContent: "updated cta text",
		},
		VideoUrl: "updated url",
	}

	newConfig, err := store.UpdateConfiguration(configuration.Id, updatedConfiguration)
	if err != nil {
		t.Fatalf("failed to update instance: %v", err)
	}
	if !equalConfigs(newConfig, expected) {
		t.Fatalf("Expected updated configuration %+v, got %+v", newConfig, expected)
	}
}
