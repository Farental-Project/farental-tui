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

	case theme.DimFontColorID:
		colorHexCode = "#374F37"

	case theme.BlurredBorderColorID, theme.BlurredFontColorID:
		colorHexCode = "#A3BBA3"

	case theme.HighlightFontColorID:
		colorHexCode = "#A3A35B"

	case theme.TitleFontColorID, theme.NormalFontColorID,
		theme.FocusedBorderColorID, theme.FocusedFontColorID:
		colorHexCode = "#132B18"

	case RacePrataarColorID:
		colorHexCode = "#056138"

	case RaceGarnothColorID:
		colorHexCode = "#AB4404"

	case BiomeMountainColorID:
		colorHexCode = "#535353"

	case BiomeFieldColorID:
		colorHexCode = "#9EA003"

	case BiomeHillColorID:
		colorHexCode = "#489B04"

	case BiomeDesertColorID:
		colorHexCode = "#96A341"

	case BiomeTropicalColorID:
		colorHexCode = "#ED9302"

	case BiomeForestColorID:
		colorHexCode = "#067501"

	case BiomeSwampColorID:
		colorHexCode = "#5F426D"

	case BiomeUndergroundColorID:
		colorHexCode = "#007C78"

	case BiomeBeachColorID:
		colorHexCode = "#F8FC05"
	}

	if colorHexCode != "" {
		color = lipgloss.Color(colorHexCode)
	}

	return color
}
