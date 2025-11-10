package theme

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn/theme"
)

type FarentalLightTheme struct {
	*FarentalDarkTheme
}

func NewFarentalLightTheme() *FarentalLightTheme {
	t := &FarentalLightTheme{
		FarentalDarkTheme: NewFarentalDarkTheme(),
	}
	t.FarentalDarkTheme.DefaultDarkTheme.Theme = t

	return t
}

func (t FarentalLightTheme) Color(id theme.ColorID) lipgloss.Color {
	var colorHexCode string

	color := t.FarentalDarkTheme.Color(id)

	switch id {
	case theme.NeutralFontColorID:
		colorHexCode = "#131313"

	case theme.NeutralDimFontColorID:
		colorHexCode = "#606060"

	case theme.BlurredBorderColorID, theme.BlurredFontColorID, theme.DimFontColorID:
		colorHexCode = "#A3BBA3"

	case theme.HighlightFontColorID:
		colorHexCode = "#A3A35B"

	case theme.TitleFontColorID, theme.NormalFontColorID,
		theme.FocusedBorderColorID, theme.FocusedFontColorID:
		colorHexCode = "#132B18"

	}

	if colorHexCode != "" {
		color = lipgloss.Color(colorHexCode)
	}

	return color
}
