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

type BaseWidget struct {
	BaseRenderable

	renderCallback func() string
}

func NewBaseWidget(renderCallback func() string) *BaseWidget {
	w := new(BaseWidget)

	w.visible = true
	w.renderCallback = renderCallback

	return w
}

func (b *BaseWidget) Init() tea.Cmd {
	return nil
}

func (b *BaseWidget) Update(msg tea.Msg) tea.Cmd {
	return nil
}
