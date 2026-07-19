package theme

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/muesli/termenv"
)

type FarentalDarkTheme struct {
	*theme.DefaultDarkTheme
}

func NewFarentalDarkTheme() *FarentalDarkTheme {
	t := &FarentalDarkTheme{
		DefaultDarkTheme: theme.NewDefaultDarkTheme(),
	}
	t.DefaultDarkTheme.Theme = t

	return t
}

func (t FarentalDarkTheme) Style(id theme.StyleID) lipgloss.Style {
	style := t.DefaultDarkTheme.Style(id)

	switch id {
	case TitleUnderlinedTextStyleID:
		style = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderTop(false).BorderLeft(false).BorderRight(false).
			Bold(true).
			Foreground(t.DefaultDarkTheme.Theme.Color(theme.TitleFontColorID)).
			BorderForeground(t.Color(theme.DimFontColorID))

	case DimUnderlinedTextStyleID:
		dfc := t.DefaultDarkTheme.Theme.Color(theme.DimFontColorID)
		style = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderTop(false).BorderLeft(false).BorderRight(false).
			Foreground(dfc).
			BorderForeground(dfc)
	}

	return style
}

func (t FarentalDarkTheme) Color(id theme.ColorID) lipgloss.Color {
	var colorHexCode string

	color := t.DefaultDarkTheme.Color(id)

	switch id {
	case theme.DimFontColorID, theme.BlurredBorderColorID, theme.BlurredFontColorID:
		// This and the unlisted-default green used for focused/normal text
		// (#18B718) are both nearest to the same standard ANSI green under
		// a reduced (ANSI/Ascii) color profile, so they become
		// indistinguishable there. Only fall back to gray in that case -
		// full-color terminals (TrueColor/ANSI256) keep the intended
		// darker green, which has plenty of resolution to stay distinct.
		if lipgloss.ColorProfile() <= termenv.ANSI256 {
			colorHexCode = "#186318"
		} else {
			colorHexCode = "#5C5C5C"
		}

	case HPBarColorID:
		colorHexCode = "#EB1F1F"

	case MPBarColorID:
		colorHexCode = "#0C67EB"

	case RacePrataarColorID:
		colorHexCode = "#05C18C"

	case RaceGarnothColorID:
		colorHexCode = "#E22D2D"

	case RaceEltrysColorID:
		colorHexCode = "#B70EF4"

	case RaceKaarColorID:
		colorHexCode = "#F4E10E"

	case RaceNymanColorID:
		colorHexCode = "#0EA8F4"

	case RaceKrynamColorID:
		colorHexCode = "#DB8C04"

	case BiomeMountainColorID:
		colorHexCode = "#939393"

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

func (t FarentalDarkTheme) Size(id theme.SizeID) int {
	size := t.DefaultDarkTheme.Size(id)

	switch id {
	case LayoutWidthSizeID:
		return 120
	}

	return size
}
