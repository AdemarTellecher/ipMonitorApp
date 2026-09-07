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
		return color.RGBA{19, 23, 31, 255} // #13171f
	case theme.ColorNameButton:
		return color.RGBA{45, 55, 72, 255} // tom neutro elegante (#2d3748) para botões secundários
	case theme.ColorNamePrimary:
		return color.RGBA{37, 99, 235, 255} // #2563eb azul vibrante para HighImportance
	case theme.ColorNameDisabled:
		return color.RGBA{70, 80, 95, 255}
	case theme.ColorNameDisabledButton:
		return color.RGBA{28, 35, 45, 255}
	case theme.ColorNameForeground:
		return color.RGBA{248, 250, 252, 255} // #f8fafc
	case theme.ColorNameInputBackground:
		return color.RGBA{16, 20, 28, 255} // #10141c
	case theme.ColorNamePlaceHolder:
		return color.RGBA{100, 116, 139, 255} // #64748b
	case theme.ColorNameSelection:
		return color.RGBA{37, 99, 235, 140}
	case theme.ColorNameHover:
		return color.RGBA{255, 255, 255, 20}
	case theme.ColorNameShadow:
		return color.RGBA{0, 0, 0, 120}
	}
	return theme.DefaultTheme().Color(n, v)
}

func (CustomDarkTheme) Size(n fyne.ThemeSizeName) float32 {
	if n == theme.SizeNamePadding {
		return 4
	}
	return theme.DefaultTheme().Size(n)
}
func (CustomDarkTheme) Font(s fyne.TextStyle) fyne.Resource     { return theme.DefaultTheme().Font(s) }
func (CustomDarkTheme) Icon(n fyne.ThemeIconName) fyne.Resource { return theme.DefaultTheme().Icon(n) }

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
