package internal

import (
	"bytes"
	"io"
	"os"
	"text/template"

	"github.com/crocoder-dev/intro-video/internal/config"
	"github.com/tdewolff/minify/v2/minify"
)

type Stylesheet struct{}

func (s Stylesheet) Process(props ProcessableFileProps, opts ProcessableFileOpts) (string, error) {

	var templatePaths []string
	var basePath string

	switch props.Theme {
	case config.DefaultTheme:
		templatePaths = append(templatePaths, "internal/template/stylesheet/default/bubble.css.tmpl")
		templatePaths = append(templatePaths, "internal/template/stylesheet/default/cta.css.tmpl")
		basePath = "internal/template/stylesheet/default/base.css"
	case config.NoneTheme:
		return "", nil
	default:
		return "", nil

	}

	t, err := template.ParseFiles(
		templatePaths...,
	)

	if err != nil {
		return "", err
	}
	file, err := os.Open(basePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var buf bytes.Buffer

	if props.Bubble.Enabled {
		err := t.ExecuteTemplate(&buf, "bubble", props.Bubble)
		if err != nil {
			return "", err
		}
	}

	if props.Cta.Enabled {
		err := t.ExecuteTemplate(&buf, "cta", props.Cta)
		if err != nil {
			return "", err
		}
	}

	base, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	var result bytes.Buffer
	result.Write(base)
	result.Write(buf.Bytes())

	var out string

	if opts.Minify {
		m := minify.Default
		out, err = m.String("text/css", result.String())
		if err != nil {
			return "", err
		}
	} else {
		out = result.String()
	}

	return out, nil
}
