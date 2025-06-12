package model

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Focusable interface {
	Focus() tea.Cmd
	Blur()
	Focused() bool
	View() string
}
