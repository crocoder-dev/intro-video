package config

import "errors"

type Theme string

const (
	DefaultTheme Theme = "default"
	ShadcnThemeLight  Theme = "shadcnLight"
	ShadcnThemeDark  Theme = "shadcnDark"
	MaterialUiThemeLight  Theme = "materialUiLight"
	MaterialUiThemeDark  Theme = "materialUiDark"
	NoneTheme    Theme = "none"
)

func NewTheme(theme string) (Theme, error) {
	switch Theme(theme) {
		case DefaultTheme, ShadcnThemeLight, ShadcnThemeDark, MaterialUiThemeDark, MaterialUiThemeLight, NoneTheme:
			return Theme(theme), nil
		default:
			return "", errors.New("invalid Theme")
	}
}
