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
	visible bool
}

func NewBaseWidget() *BaseWidget {
	w := new(BaseWidget)

	w.visible = true

	return w
}

func (b *BaseWidget) Init() tea.Cmd {
	return nil
}

func (b *BaseWidget) Update(msg tea.Msg) tea.Cmd {
	return nil
}

func (b *BaseWidget) Resize(size Size) {}

func (b *BaseWidget) GetSize() Size {
	return NewSize(0, 0)
}

func (b *BaseWidget) GetMinSize() Size {
	return NewSize(0, 0)
}

func (b *BaseWidget) GetPreferredSize() Size {
	return NewSize(0, 0)
}

func (b *BaseWidget) GetMaxSize() Size {
	return NewSize(0, 0)
}

func (b *BaseWidget) SetVisible(visible bool) {
	b.visible = visible
}

func (b *BaseWidget) IsVisible() bool {
	return b.visible
}
