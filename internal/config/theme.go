package config

import "errors"

type Theme string

const (
	DefaultTheme Theme = "default"
	ShadcnThemeLight  Theme = "shadcnLight"
	ShadcnThemeDark  Theme = "shadcnDark"
	NoneTheme    Theme = "none"
)

func NewTheme(theme string) (Theme, error) {
	switch Theme(theme) {
		case DefaultTheme:
			return Theme(theme), nil
		case ShadcnThemeLight:
			return Theme(theme), nil
		case ShadcnThemeDark:
			return Theme(theme), nil
		case NoneTheme:
			return Theme(theme), nil
		default:
			return "", errors.New("invalid Theme")
	}
}
