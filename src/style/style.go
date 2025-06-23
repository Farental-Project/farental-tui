package style

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	FocusedStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(ColorSpecialHighlight))

	BlurredStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(ColorHighlightDim))

	NormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorHighlight))

	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorHighlight)).
			Bold(true)

	DimTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorHighlightDim)).
			Italic(true)

	NeutralLessDimTextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorNeutralLessDim)).
				Italic(true)

	NeutralDimTextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorNeutralDim)).
				Italic(true)

	BoldTextStyle = lipgloss.NewStyle().Bold(true)

	ItalicTextStyle = lipgloss.NewStyle().Italic(true)

	ContainerStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(ColorHighlight))

	ContainerTitleStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color(ColorHighlightDim)).
				Foreground(lipgloss.Color(ColorHighlightDim)).
				Italic(true).
				BorderTop(false).BorderRight(false).BorderLeft(false)

	DimBottomBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color(ColorHighlightDim)).
				BorderTop(false).BorderRight(false).BorderLeft(false)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorError)).
			Bold(true)

	TextStyle = lipgloss.NewStyle()

	HighlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorSpecialHighlight)).
			Bold(true)
)
