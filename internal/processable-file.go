package internal

import "github.com/crocoder-dev/intro-video/internal/config"

type ProcessableFileProps struct {
	config.Theme
	config.Bubble
	config.Cta
	URL string
}

type ProcessableFileOpts struct {
	Preview bool
	Minify bool
}

type ProcessableFile interface {
	Process(props ProcessableFileProps, opts ProcessableFileOpts) (string, error)
}
