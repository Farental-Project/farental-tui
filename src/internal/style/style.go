package style

import (
	ftheme "farental/internal/theme"

	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
)

var (
	MainHelpStyle bubblehelp.Style
)

func InitHelpStyle() {
	t := orvyn.GetTheme()
	ds := t.Style(theme.NeutralDimTextStyleID)
	ns := t.Style(theme.NeutralTextStyleID)

	MainHelpStyle = bubblehelp.Style{
		EssentialKey:               ns.Bold(true),
		EssentialKeyDescription:    ds,
		EssentialKeySeparator:      ds,
		EssentialKeySeparatorValue: " ",
		EssentialColSeparator:      ds,
		EssentialColSeparatorValue: " • ",
		FullKey:                    ns.Bold(true),
		FullKeyDescription:         ds,
		FullKeySeparator:           ds,
		FullKeySeparatorValue:      " ",
		FullColSeparator:           lipgloss.Style{},
		FullColSeparatorValue:      "  ",
	}
}

func RaceStyle(name string) lipgloss.Style {
	var style lipgloss.Style
	var colorID theme.ColorID

	style = lipgloss.NewStyle()
	t := orvyn.GetTheme()

	switch name {
	case "Prataar":
		colorID = ftheme.RacePrataarColorID
	case "Garnoth":
		colorID = ftheme.RaceGarnothColorID
	default:
		colorID = theme.NormalFontColorID
	}

	style = style.Foreground(t.Color(colorID))

	return style
}

func LocationBiomeStyle(code string) lipgloss.Style {
	var style lipgloss.Style
	var colorID theme.ColorID

	style = lipgloss.NewStyle()
	t := orvyn.GetTheme()

	switch code {
	case "mountain":
		colorID = ftheme.BiomeMountainColorID
	case "field":
		colorID = ftheme.BiomeFieldColorID
	case "hill":
		colorID = ftheme.BiomeHillColorID
	case "desert":
		colorID = ftheme.BiomeDesertColorID
	case "tropical":
		colorID = ftheme.BiomeTropicalColorID
	case "forest":
		colorID = ftheme.BiomeForestColorID
	case "swamp":
		colorID = ftheme.BiomeSwampColorID
	case "underground":
		colorID = ftheme.BiomeUndergroundColorID
	case "beach":
		colorID = ftheme.BiomeBeachColorID
	default:
		colorID = theme.NormalFontColorID
	}

	style = style.Foreground(t.Color(colorID))

	return style
}
