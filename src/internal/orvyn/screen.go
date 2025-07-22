package orvyn

import (
	tea "github.com/charmbracelet/bubbletea"
)

type ScreenID string

// Screen defines behaviour of an Orvyn screen.
type Screen interface {
	// OnEnter is called when the screen is entered. Can take as parameter the struct from the previous screen.
	OnEnter(interface{}) tea.Cmd

	// OnExit is called when the screen is being exited. Can return a struct that will be passed to the next screen.
	OnExit() interface{}

	// Updatable Screen can be updated.
	Updatable

	// Render returns the view string of the whole screen
	Render() Layout

	// GetID returns the ScreenID of the screen.
	GetID() ScreenID
}

type BaseScreen struct {
	ID      ScreenID
	Widgets []Widget
}

func (b *BaseScreen) GetID() ScreenID {
	return b.ID
}

// OnEnter default behaviour init every Widget.
func (b *BaseScreen) OnEnter(_ interface{}) tea.Cmd {
	var cmds []tea.Cmd

	for _, w := range b.Widgets {
		cmds = append(cmds, w.Init())
	}

	return tea.Batch(cmds...)
}

// Update default behaviour updates every Widget.
func (b *BaseScreen) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	for _, w := range b.Widgets {
		cmds = append(cmds, w.Update(msg))
	}

	return tea.Batch(cmds...)
}

func (b *BaseScreen) AddWidget(w Widget) {
	b.Widgets = append(b.Widgets, w)
}
