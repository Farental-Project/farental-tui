package model

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	focusedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#39d800")).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#39d800")).
			Padding(0, 1)

	blurredStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1b6800")).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#1b6800")).
			Padding(0, 1)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#39d800")).
			Bold(true)

	containerStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("#39d800")).
			Padding(2, 4).
			Margin(1, 2)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#dd7302")).
			Bold(true)
)
