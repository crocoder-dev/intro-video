package template

// This is a dummy file to make sure `go mod tidy` does not remove dependency a-h/templ

import (
	_ "github.com/a-h/templ"
)

var DO_NOT_DELETE = "this file"
