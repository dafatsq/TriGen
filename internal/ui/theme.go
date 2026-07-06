package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type StudioTheme struct{}

func (m *StudioTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// We customize colors for a sleek, Slate & Cyan dark-mode experience.
	switch name {
	case theme.ColorNameBackground:
		return color.RGBA{R: 0x0f, G: 0x17, B: 0x2a, A: 0xff} // Slate 900
	case theme.ColorNameInputBackground:
		return color.RGBA{R: 0x12, G: 0x18, B: 0x24, A: 0xff} // Deep slate input box
	case theme.ColorNameForeground:
		return color.RGBA{R: 0xf8, G: 0xfa, B: 0xfc, A: 0xff} // Slate 50 (white/grey text)
	case theme.ColorNamePrimary:
		return color.RGBA{R: 0x14, G: 0xb8, B: 0xa6, A: 0xff} // Teal 500 primary accent
	case theme.ColorNameSuccess:
		return color.RGBA{R: 0x10, G: 0xb9, B: 0x81, A: 0xff} // Emerald 500
	case theme.ColorNameError:
		return color.RGBA{R: 0xef, G: 0x44, B: 0x44, A: 0xff} // Red 500
	case theme.ColorNameWarning:
		return color.RGBA{R: 0xf5, G: 0x9e, B: 0x0b, A: 0xff} // Amber 500
	case theme.ColorNameHover:
		return color.RGBA{R: 0x33, G: 0x41, B: 0x55, A: 0xff} // Slate 700
	case theme.ColorNamePressed:
		return color.RGBA{R: 0x1e, G: 0x29, B: 0x3b, A: 0xff}
	case theme.ColorNameDisabled:
		return color.RGBA{R: 0x47, G: 0x55, B: 0x69, A: 0xff} // Slate 600
	case theme.ColorNamePlaceHolder:
		return color.RGBA{R: 0x64, G: 0x74, B: 0x8b, A: 0xff} // Slate 500
	case theme.ColorNameScrollBar:
		return color.RGBA{R: 0x47, G: 0x55, B: 0x69, A: 0xaa} // Transparent slate
	case theme.ColorNameSelection:
		return color.RGBA{R: 0x14, G: 0xb8, B: 0xa6, A: 0x33} // Transparent Teal
	case theme.ColorNameFocus:
		return color.RGBA{R: 0x14, G: 0xb8, B: 0xa6, A: 0x77} // Focus border teal
	case theme.ColorNameHeaderBackground:
		return color.RGBA{R: 0x1e, G: 0x29, B: 0x3b, A: 0xff}
	case theme.ColorNameButton:
		return color.RGBA{R: 0x1e, G: 0x29, B: 0x3b, A: 0xff}
	}
	return theme.DefaultTheme().Color(name, theme.VariantDark)
}

func (m *StudioTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (m *StudioTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m *StudioTheme) Size(name fyne.ThemeSizeName) float32 {
	// Slightly increase padding/spacing for a modern spacious layout.
	if name == theme.SizeNamePadding {
		return 8
	}
	if name == theme.SizeNameInputBorder {
		return 1.5
	}
	return theme.DefaultTheme().Size(name)
}
