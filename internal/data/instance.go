package data

import (
	"database/sql"
	"os"

	"github.com/crocoder-dev/intro-video/internal/config"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type Instance struct {
	Id             int32
	Uuid           []byte
	Videos         map[int32]Video
	Configurations map[int32]Configuration
}

type NewVideo struct {
	Weight int32
	URL    string
}

type Video struct {
	Id              int32
	Weight          int32
	ConfigurationId int32
	URL             string
}

type NewConfiguration struct {
	Bubble config.Bubble
	Cta    config.Cta
}

type Configuration struct {
	Id     int32
	Bubble config.Bubble
	Cta    config.Cta
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
func (s *Store) LoadInstance(uuid []byte) (Instance, error) {
	db, err := sql.Open(s.DriverName, s.DatabaseUrl)
	if err != nil {
		return Instance{}, err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return Instance{}, err
	}

	rows, err := tx.Query(`
		SELECT
		videos.id,
		videos.weight,
		videos.url,
		videos.configuration_id
		FROM instances
		JOIN videos ON videos.instance_id = instances.id
		WHERE instances.uuid = ?;
		`,
		uuid,
	)

	if err != nil {
		tx.Rollback()
		return Instance{}, err
	}
	defer rows.Close()

	instance := Instance{Uuid: uuid, Videos: map[int32]Video{}, Configurations: map[int32]Configuration{}}

	for rows.Next() {
		var video Video

		if err := rows.Scan(
			&video.Id,
			&video.Weight,
			&video.URL,
			&video.ConfigurationId,
		); err != nil {
			tx.Rollback()
			return Instance{}, err
		}

		instance.Videos[video.Id] = video
	}

	rows, err = tx.Query(`
		SELECT DISTINCT
		config.id,
		config.bubble_enabled,
		config.bubble_text_content,
		config.bubble_type,
		config.cta_enabled,
		config.cta_text_content,
		config.cta_type
		FROM instances
		JOIN videos ON videos.instance_id = instances.id
		JOIN configurations as config ON videos.configuration_id = config.id
		WHERE instances.uuid = ?;
		`,
		uuid,
	)
	if err != nil {
		tx.Rollback()
		return Instance{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var configuration Configuration

		configuration.Bubble = config.Bubble{}
		configuration.Cta = config.Cta{}

		if err := rows.Scan(
			&configuration.Id,
			&configuration.Bubble.Enabled,
			&configuration.Bubble.TextContent,
			&configuration.Bubble.Type,
			&configuration.Cta.Enabled,
			&configuration.Cta.TextContent,
			&configuration.Cta.Type,
		); err != nil {
			tx.Rollback()
			return Instance{}, err
		}
		instance.Configurations[configuration.Id] = configuration
	}

	err = tx.Commit()
	if err != nil {
		return Instance{}, err
	}

	return instance, nil
}

func (s *Store) CreateInstance(video NewVideo, configuration NewConfiguration) (Instance, error) {
	db, err := sql.Open(s.DriverName, s.DatabaseUrl)
	if err != nil {
		return Instance{}, err

	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return Instance{}, err
	}

	newUUID := uuid.New()

	binUUID, err := newUUID.MarshalBinary()

	var instanceId int32

	err = tx.QueryRow(`
		INSERT INTO instances (uuid)
		VALUES (?)
		RETURNING id;
	`, binUUID).Scan(&instanceId)

	if err != nil {
		tx.Rollback()
		return Instance{}, err
	}

	var configurationId int32

	err = tx.QueryRow(`
		INSERT INTO configurations
		(
			bubble_enabled,
			bubble_text_content,
			bubble_type,
			cta_enabled,
			cta_text_content,
			cta_type
		)
		VALUES (?, ?, ?, ?, ?, ?)
		RETURNING id;
		`,
		configuration.Bubble.Enabled,
		configuration.Bubble.TextContent,
		configuration.Bubble.Type,
		configuration.Cta.Enabled,
		configuration.Cta.TextContent,
		configuration.Cta.Type,
	).Scan(&configurationId)

	if err != nil {
		tx.Rollback()
		return Instance{}, err
	}

	var videoId int32

	err = tx.QueryRow(`
		INSERT INTO videos
		(
			weight,
			url,
			configuration_id,
			instance_id
		)
		Values (?, ?, ?)
		RETURNING id;
		`,
		video.Weight,
		video.URL,
		configurationId,
		instanceId,
	).Scan(&videoId)

	err = tx.Commit()
	if err != nil {
		return Instance{}, err
	}

	instance := Instance{
		Id:             instanceId,
		Uuid:           binUUID,
		Videos:         map[int32]Video{},
		Configurations: map[int32]Configuration{},
	}

	instance.Videos[videoId] = Video{
		Id:              videoId,
		Weight:          video.Weight,
		URL:             video.URL,
		ConfigurationId: configurationId,
	}

	instance.Configurations[configurationId] = Configuration{
		Id:     configurationId,
		Bubble: configuration.Bubble,
		Cta:    configuration.Cta,
	}

	return instance, nil
}