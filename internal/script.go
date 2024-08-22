package internal

import (
	"bytes"
	"errors"
	"html/template"
	"io"
	"os"

	"github.com/tdewolff/minify/v2/minify"
)

type Script struct{}

func (s Script) Process(props ProcessableFileProps, opts ProcessableFileOpts) (string, error) {
	if props.URL == "" {
		return "", errors.New("video URL is required")
	}

	t, err := template.ParseFiles(
		"internal/template/script/start.js.tmpl",
		"internal/template/script/start.preview.js.tmpl",
		"internal/template/script/end.js.tmpl",
		"internal/template/script/end.preview.js.tmpl",
		"internal/template/script/video.js.tmpl",
		"internal/template/script/bubble.js.tmpl",
		"internal/template/script/cta.js.tmpl",
	)

	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	var end bytes.Buffer

	if opts.Preview {
		err = t.ExecuteTemplate(&buf, "start-preview", nil)
	} else {
		err = t.ExecuteTemplate(&buf, "start", props.Bubble)
	}

	if err != nil {
		return "", err
	}

	err = t.ExecuteTemplate(&buf, "video", props)
	if err != nil {
		return "", err
	}

	err = t.ExecuteTemplate(&buf, "bubble", props.Bubble)
	if err != nil {
		return "", err
	}

	err = t.ExecuteTemplate(&buf, "cta", props.Cta)
	if err != nil {
		return "", err
	}

	if opts.Preview {
		err = t.ExecuteTemplate(&end, "end-preview", nil)
	} else {
		err = t.ExecuteTemplate(&end, "end", nil)
	}

	if err != nil {
		return "", err
	}
	var file *os.File
	var base []byte

	if !opts.Preview {
		file, err := os.Open("internal/template/script/base.js")
		if err != nil {
			return "", err
		}
		defer file.Close()
		base, err = io.ReadAll(file)
		if err != nil {
			return "", err
		}
	}

	var run []byte
	var previewScript []byte

	if !opts.Preview {
		file, err = os.Open("internal/template/script/run.js")
		if err != nil {
			return "", err
		}
		defer file.Close()

		run, err = io.ReadAll(file)
		if err != nil {
			return "", err
		}
	} else {
		file, err = os.Open("internal/template/script/preview.js")
		if err != nil {
			return "", err
		}
		defer file.Close()

		previewScript, err = io.ReadAll(file)
		if err != nil {
			return "", err
		}
	}

	var result bytes.Buffer
	result.Write(buf.Bytes())
	result.Write(base)
	result.Write(run)
	result.Write(previewScript)
	result.Write(end.Bytes())

	var out string

	if opts.Minify {
		m := minify.Default
		out, err = m.String("text/javascript", result.String())
		if err != nil {
			return "", err
		}
	} else {
		out = result.String()
	}

	return out, nil
}
