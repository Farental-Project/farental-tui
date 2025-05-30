package style

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	FocusedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorHighlight)).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(ColorHighlight)).
			Padding(0, 1)

	BlurredStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorHighlightDim)).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(ColorHighlightDim)).
			Padding(0, 1)

	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorHighlight)).
			Bold(true)

	DimTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorHighlightDim)).
			Italic(true)

	NeutralDimTextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorNeutralDim)).
				Italic(true)

	BoldTextStyle = lipgloss.NewStyle().Bold(true)

	ItalicTextStyle = lipgloss.NewStyle().Italic(true)

	ContainerStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(ColorHighlight))

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorError)).
			Bold(true)

	TextStyle = lipgloss.NewStyle()

	HighlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorSpecialHighlight)).
			Bold(true)

	ContainerTitleStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color(ColorHighlightDim)).
				Foreground(lipgloss.Color(ColorHighlightDim)).
				Italic(true).
				BorderTop(false).BorderRight(false).BorderLeft(false)
)
