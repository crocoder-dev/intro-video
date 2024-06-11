package data

import (
	"database/sql"
	"os"

	"github.com/crocoder-dev/intro-video/internal"
	"github.com/crocoder-dev/intro-video/internal/config"
	"github.com/joho/godotenv"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type Video struct {
	Id              int32
	Weight          int32
	URL             string
	ConfigurationId int32
	internal.ProcessableFileProps
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

func (s *Store) LoadInstance(id int32) (map[int32]Video, error) {
	db, err := sql.Open(s.DriverName, s.DatabaseUrl)
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(`
		SELECT
			videos.id,
			videos.weight,
			videos.url,
			confs.id,
			confs.bubble_enabled,
			confs.bubble_text_content,
			confs.bubble_type,
			confs.cta_enabled,
			confs.cta_text_content,
			confs.cta_type
		FROM instances
		JOIN videos ON videos.instance_id = instances.id
		JOIN configurations as confs ON confs.id = videos.configuration_id
		WHERE instances.id = $1;
		`,
		id,
	)
	defer rows.Close()

	videos := make(map[int32]Video)

	for rows.Next() {
		var video Video
		video.Bubble = config.Bubble{}
		video.Cta = config.Cta{}

		if err := rows.Scan(
			&video.Id,
			&video.Weight,
			&video.URL,
			&video.ConfigurationId,
			&video.Bubble.Enabled,
			&video.Bubble.TextContent,
			&video.Bubble.Type,
			&video.Cta.Enabled,
			&video.Cta.TextContent,
			&video.Cta.Type,
		); err != nil {
			return nil, err
		}

		videos[video.Id] = video
	}


	return videos, nil
}

func (s *Store) SaveInstance(id int32, instance map[int32]Video) error {
	return nil
}
