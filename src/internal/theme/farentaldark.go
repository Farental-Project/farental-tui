package theme

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn/theme"
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

func (t FarentalDarkTheme) Size(id theme.SizeID) int {
	size := t.DefaultDarkTheme.Size(id)

	switch id {
	case LayoutWidthSizeID:
		return 120
	}

	return size
}
