package orvyn

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Widget interface {
	// Init is called on the Widget when entering a Screen that contains it.
	Init() tea.Cmd

	// Updatable Widget can be updated.
	Updatable

	// Renderable Widget can be rendered.
	Renderable
}
