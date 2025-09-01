package theme

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn/theme"
)

const (
	TitleUnderlinedTextStyleID theme.StyleID = iota
	DimUnderlinedTextStyleID
)

const (
	HPBarColorID theme.ColorID = iota
	MPBarColorID
)

const (
	LayoutWidthSizeID theme.SizeID = iota
)

type FarentalTheme struct {
	theme.DefaultDarkTheme
}

func (t FarentalTheme) Style(id theme.StyleID) lipgloss.Style {
	style := t.DefaultDarkTheme.Style(id)

	switch id {

	case TitleUnderlinedTextStyleID:
		style = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderTop(false).BorderLeft(false).BorderRight(false).
			Bold(true).
			Foreground(t.Color(theme.TitleFontColorID))

	case DimUnderlinedTextStyleID:
		style = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderTop(false).BorderLeft(false).BorderRight(false).
			Foreground(t.Color(theme.DimFontColorID))
	}

	return style
}

func (t FarentalTheme) Color(id theme.ColorID) lipgloss.Color {
	var colorHexCode string

	color := t.DefaultDarkTheme.Color(id)

	switch id {
	case HPBarColorID:
		colorHexCode = "#EB1F1F"

	case MPBarColorID:
		colorHexCode = "#0C67EB"

	}

	if colorHexCode != "" {
		color = lipgloss.Color(colorHexCode)
	}

	return color
}

func (t FarentalTheme) Size(id theme.SizeID) int {
	size := t.DefaultDarkTheme.Size(id)

	switch id {
	case LayoutWidthSizeID:
		return 120
	}

	return size
}
