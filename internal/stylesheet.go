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
	case config.ShadcnThemeLight:
		templatePaths = append(templatePaths, "internal/template/stylesheet/shadcn/light/bubble.css.tmpl")
		templatePaths = append(templatePaths, "internal/template/stylesheet/shadcn/light/cta.css.tmpl")
		basePath = "internal/template/stylesheet/shadcn/light/base.css"
	case config.ShadcnThemeDark:
		templatePaths = append(templatePaths, "internal/template/stylesheet/shadcn/dark/bubble.css.tmpl")
		templatePaths = append(templatePaths, "internal/template/stylesheet/shadcn/dark/cta.css.tmpl")
		basePath = "internal/template/stylesheet/shadcn/dark/base.css"
	case config.MaterialUiThemeLight:
		templatePaths = append(templatePaths, "internal/template/stylesheet/material/light/bubble.css.tmpl")
		templatePaths = append(templatePaths, "internal/template/stylesheet/material/light/cta.css.tmpl")
		basePath = "internal/template/stylesheet/material/light/base.css"
	case config.MaterialUiThemeDark:
		templatePaths = append(templatePaths, "internal/template/stylesheet/material/dark/bubble.css.tmpl")
		templatePaths = append(templatePaths, "internal/template/stylesheet/material/dark/cta.css.tmpl")
		basePath = "internal/template/stylesheet/material/dark/base.css"
	case config.TailwindThemeDark:
		templatePaths = append(templatePaths, "internal/template/stylesheet/tailwind/dark/bubble.css.tmpl")
		templatePaths = append(templatePaths, "internal/template/stylesheet/tailwind/dark/cta.css.tmpl")
		basePath = "internal/template/stylesheet/tailwind/dark/base.css"
	case config.TailwindThemeLight:
		templatePaths = append(templatePaths, "internal/template/stylesheet/tailwind/light/bubble.css.tmpl")
		templatePaths = append(templatePaths, "internal/template/stylesheet/tailwind/light/cta.css.tmpl")
		basePath = "internal/template/stylesheet/tailwind/light/base.css"
	case config.Crocoder:
		templatePaths = append(templatePaths, "internal/template/stylesheet/crocoder/bubble.css.tmpl")
		templatePaths = append(templatePaths, "internal/template/stylesheet/crocoder/cta.css.tmpl")
		basePath = "internal/template/stylesheet/crocoder/base.css"
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
