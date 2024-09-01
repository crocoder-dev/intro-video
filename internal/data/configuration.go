package data

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/crocoder-dev/intro-video/internal/config"
	"github.com/joho/godotenv"
	"github.com/oklog/ulid/v2"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type Configuration struct {
	Id       []byte
	Theme    config.Theme
	Bubble   config.Bubble
	Cta      config.Cta
	VideoUrl string
}

type Store struct {
	DatabaseUrl string
	DriverName  string
}

func NewStore() (Store, error) {
	err := godotenv.Load(".env")
	if err != nil {
		return Store{}, err
	}
	url := os.Getenv("DATABASE_URL")

	return Store{DatabaseUrl: url, DriverName: "libsql"}, nil

}

func (s *Store) LoadConfiguration(id []byte) (Configuration, error) {
	db, err := sql.Open(s.DriverName, s.DatabaseUrl)
	if err != nil {
		return Configuration{}, err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return Configuration{}, err
	}
	defer tx.Rollback()

	var (
		configId          []byte
		videoURL          string
		theme             string
		ctaEnabled        bool
		ctaTextContent    string
		bubbleEnabled     bool
		bubbleTextContent string
	)

	err = tx.QueryRow(`
		SELECT
			id,
			video_url,
			theme,
			cta_enabled,
			cta_text_content,
			bubble_enabled,
			bubble_text_content
		FROM configurations
		WHERE id = ?
	`, id).Scan(
		&configId,
		&videoURL,
		&theme,
		&ctaEnabled,
		&ctaTextContent,
		&bubbleEnabled,
		&bubbleTextContent,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return Configuration{}, fmt.Errorf("no configuration found with id %x", id)
		}
		return Configuration{}, err
	}

	return Configuration{
		Id:    configId,
		Theme: config.Theme(theme),
		Bubble: config.Bubble{
			Enabled:     bubbleEnabled,
			TextContent: bubbleTextContent,
		},
		Cta: config.Cta{
			Enabled:     ctaEnabled,
			TextContent: ctaTextContent,
		},
		VideoUrl: videoURL,
	}, nil
}

func (s *Store) CreateConfiguration(configuration Configuration) (Configuration, error) {
	db, err := sql.Open(s.DriverName, s.DatabaseUrl)
	if err != nil {
		return Configuration{}, err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return Configuration{}, err
	}

	var configurationId []byte

	newUlid := ulid.Make()

	binUlid := newUlid.Bytes()

	err = tx.QueryRow(`
		INSERT INTO configurations
		(
			id,
			video_url,
			theme,
			bubble_enabled,
			bubble_text_content,
			cta_enabled,
			cta_text_content
		)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		RETURNING id;
		`,
		binUlid,
		configuration.VideoUrl,
		configuration.Theme,
		configuration.Bubble.Enabled,
		configuration.Bubble.TextContent,
		configuration.Cta.Enabled,
		configuration.Cta.TextContent,
	).Scan(&configurationId)

	if err != nil {
		tx.Rollback()
		return Configuration{}, err
	}

	err = tx.Commit()
	if err != nil {
		return Configuration{}, err
	}

	newConfiguration := Configuration{
		Id:       configurationId,
		Theme:    configuration.Theme,
		Bubble:   config.Bubble{Enabled: configuration.Bubble.Enabled, TextContent: configuration.Bubble.TextContent},
		Cta:      config.Cta{Enabled: configuration.Cta.Enabled, TextContent: configuration.Cta.TextContent},
		VideoUrl: configuration.VideoUrl,
	}

	return newConfiguration, nil
}

func (s *Store) UpdateConfiguration(id []byte, configuration Configuration) (Configuration, error) {
	db, err := sql.Open(s.DriverName, s.DatabaseUrl)
	if err != nil {
		return Configuration{}, err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return Configuration{}, err
	}

	_, err = tx.Exec(`
		UPDATE configurations
		SET
			video_url = ?,
			theme = ?,
			bubble_enabled = ?,
			bubble_text_content = ?,
			cta_enabled = ?,
			cta_text_content = ?
		WHERE id = ?
		`,
		configuration.VideoUrl,
		configuration.Theme,
		configuration.Bubble.Enabled,
		configuration.Bubble.TextContent,
		configuration.Cta.Enabled,
		configuration.Cta.TextContent,
		id,
	)

	if err != nil {
		tx.Rollback()
		return Configuration{}, err
	}

	err = tx.Commit()
	if err != nil {
		return Configuration{}, err
	}

	updatedConfiguration := Configuration{
		Id:       id,
		Theme:    configuration.Theme,
		Bubble:   config.Bubble{Enabled: configuration.Bubble.Enabled, TextContent: configuration.Bubble.TextContent},
		Cta:      config.Cta{Enabled: configuration.Cta.Enabled, TextContent: configuration.Cta.TextContent},
		VideoUrl: configuration.VideoUrl,
	}

	return updatedConfiguration, nil
}
