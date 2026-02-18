package view

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type CustomDarkTheme struct{}

func (CustomDarkTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	switch n {
	case theme.ColorNameBackground:
		return color.RGBA{30, 32, 40, 255}
	case theme.ColorNameButton:
		return color.RGBA{50, 50, 60, 255}
	case theme.ColorNameDisabled:
		return color.RGBA{80, 80, 80, 255}
	case theme.ColorNameForeground:
		return color.White
	case theme.ColorNameInputBackground:
		return color.RGBA{50, 50, 60, 255}
	case theme.ColorNamePlaceHolder:
		return color.RGBA{120, 120, 120, 255}
	case theme.ColorNameSelection:
		return color.RGBA{100, 150, 255, 255}
	}
	return theme.DefaultTheme().Color(n, v)
}
func (CustomDarkTheme) Font(s fyne.TextStyle) fyne.Resource     { return theme.DefaultTheme().Font(s) }
func (CustomDarkTheme) Icon(n fyne.ThemeIconName) fyne.Resource { return theme.DefaultTheme().Icon(n) }
func (CustomDarkTheme) Size(n fyne.ThemeSizeName) float32       { return theme.DefaultTheme().Size(n) }

type CustomLightTheme struct{}

func (CustomLightTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	switch n {
	case theme.ColorNameBackground:
		return color.RGBA{245, 247, 255, 255}
	case theme.ColorNameButton:
		return color.RGBA{230, 230, 240, 255}
	case theme.ColorNameDisabled:
		return color.RGBA{200, 200, 200, 255}
	case theme.ColorNameForeground:
		return color.Black
	case theme.ColorNameInputBackground:
		return color.White
	case theme.ColorNamePlaceHolder:
		return color.RGBA{150, 150, 150, 255}
	case theme.ColorNameSelection:
		return color.RGBA{100, 150, 255, 255}
	}
	return theme.DefaultTheme().Color(n, v)
}
func (CustomLightTheme) Font(s fyne.TextStyle) fyne.Resource     { return theme.DefaultTheme().Font(s) }
func (CustomLightTheme) Icon(n fyne.ThemeIconName) fyne.Resource { return theme.DefaultTheme().Icon(n) }
func (CustomLightTheme) Size(n fyne.ThemeSizeName) float32       { return theme.DefaultTheme().Size(n) }
