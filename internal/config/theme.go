package config

import "errors"

type Theme string

const (
	DefaultTheme Theme = "default"
	NoneTheme    Theme = "none"
)

func NewTheme(theme string) (Theme, error) {
	switch Theme(theme) {
		case DefaultTheme:
			return Theme(theme), nil
		case NoneTheme:
			return Theme(theme), nil
		default:
			return "", errors.New("invalid Theme")
	}
}
