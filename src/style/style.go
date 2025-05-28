package style

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	FocusedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#39d800")).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#39d800")).
			Padding(0, 1)

	BlurredStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1b6800")).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#1b6800")).
			Padding(0, 1)

	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#39d800")).
			Bold(true)

	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1c490c")).
			Italic(true)

	ContainerStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("#39d800")).
			Padding(2, 4).
			Margin(1, 2)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#dd7302")).
			Bold(true)

	HighlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#75f902")).
			Bold(true)
)
