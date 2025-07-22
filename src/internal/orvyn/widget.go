package orvyn

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Widget interface {
	// Init is called on the Widget when entering a Screen that contains it.
	Init() tea.Cmd

	// Update manages messages at a Widget level.
	Update(tea.Msg) tea.Cmd

	// Render defines the visual of the widget.
	// Gets a *Size as parameters when the Widget is rendered from a Layout. Can be nil.
	Render(*Size) string
}
