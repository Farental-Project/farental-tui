package style

import (
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
	var color string

	style = lipgloss.NewStyle()

	switch name {
	case "Prataar":
		color = "#05c18c"
	case "Garnoth":
		color = "#db8c04"
	default:
		color = ""
	}

	style = style.Foreground(lipgloss.Color(color))

	return style
}

func LocationBiomeStyle(code string) lipgloss.Style {
	var style lipgloss.Style
	var color string

	style = lipgloss.NewStyle()

	switch code {
	case "mountain":
		color = "#939393"
	case "field":
		color = "#9ea003"
	case "hill":
		color = "#489b04"
	case "desert":
		color = "#96a341"
	case "tropical":
		color = "#ed9302"
	case "forest":
		color = "#067501"
	case "swamp":
		color = "#5f426d"
	case "underground":
		color = "#007c78"
	case "beach":
		color = "#f8fc05"
	default:
		color = ""
	}

	style = style.Foreground(lipgloss.Color(color))

	return style
}
