package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type AppTheme struct{}

var _ fyne.Theme = (*AppTheme)(nil)

// Colors
var (
	// Palette
	accentColor = color.RGBA{R: 0x2F, G: 0x78, B: 0x6F, A: 0xFF} // Accent (#2F786F)
	primaryGray = color.RGBA{R: 0x75, G: 0x75, B: 0x75, A: 0xFF} // Primary (Mid Gray)
	buttonGray  = color.RGBA{R: 0x42, G: 0x42, B: 0x42, A: 0xFF} // Button Normal (Dark Gray)
	bgGray      = color.RGBA{R: 0x1E, G: 0x1E, B: 0x1E, A: 0xFF} // Background (from Logo #1E1E1E)
	surfaceGray = color.RGBA{R: 0x2A, G: 0x2A, B: 0x2A, A: 0xFF} // Surface/Input (Slightly lighter than BG)
	grey200     = color.RGBA{R: 0xEE, G: 0xEE, B: 0xEE, A: 0xFF} // Text
	grey400     = color.RGBA{R: 0xBD, G: 0xBD, B: 0xBD, A: 0xFF} // Disabled/Placeholder

	// State Colors
	hoverColor     = accentColor                                     // Highlight on hover
	pressedColor   = color.RGBA{R: 0x2F, G: 0x78, B: 0x6F, A: 0xCC}  // Accent color semi-transparent on click
	scrollBarColor = color.NRGBA{R: 0x75, G: 0x75, B: 0x75, A: 0x99} // Primary transparent
	shadowColor    = color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x66} // Black shadow
	focusColor     = accentColor                                     // Focus highlight
	selectionColor = accentColor                                     // Selection highlight
)

func (t AppTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return bgGray
	case theme.ColorNameButton:
		return buttonGray
	case theme.ColorNameDisabledButton:
		return surfaceGray
	case theme.ColorNameDisabled:
		return grey400
	case theme.ColorNameForeground:
		return grey200
	case theme.ColorNameHover:
		return hoverColor
	case theme.ColorNameInputBackground:
		return surfaceGray
	case theme.ColorNamePlaceHolder:
		return grey400
	case theme.ColorNamePressed:
		return pressedColor
	case theme.ColorNamePrimary:
		return primaryGray
	case theme.ColorNameScrollBar:
		return scrollBarColor
	case theme.ColorNameShadow:
		return shadowColor
	case theme.ColorNameFocus:
		return focusColor
	case theme.ColorNameSelection:
		return selectionColor
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (t AppTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t AppTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t AppTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

// TreeTheme overrides AppTheme specifically for tree widget to enforce custom hover/selection colors
type TreeTheme struct {
	fyne.Theme
}

func (t TreeTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameHover:
		return accentColor
	case theme.ColorNameSelection:
		return primaryGray
	default:
		return t.Theme.Color(name, variant)
	}
}
