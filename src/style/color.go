package style

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	ColorHighlight        = "#39d800"
	ColorHighlightDim     = "#389118"
	ColorNeutralLessDim   = "#999999"
	ColorNeutralDim       = "#707070"
	ColorError            = "#dd7302"
	ColorSpecialHighlight = "#75f902"
)

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
