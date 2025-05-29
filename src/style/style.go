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

	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorDim)).
			Italic(true)

	ContainerStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(ColorHighlight))

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorError)).
			Bold(true)

	HighlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorSpecialHighlight)).
			Bold(true)
)
