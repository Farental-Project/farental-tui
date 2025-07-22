package orvyn

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Layout interface {
	Renderable
	Updatable

	ResizeLayout(*Size)
	SetView(string)
	GetElements() []Renderable
	IsDirty() bool
	SetDirty(bool)
}

type BaseLayout struct {
	elements []Renderable

	dirty bool
	view  string
}

func NewBaseLayout(elements []Renderable) *BaseLayout {
	b := new(BaseLayout)

	b.elements = elements
	b.dirty = true

	return b
}

func (b *BaseLayout) ResizeLayout(size *Size) {}

func (b *BaseLayout) Update(msg tea.Msg) tea.Cmd {
	switch msg.(type) {
	case tea.WindowSizeMsg:
		b.SetDirty(true)
	}

	return nil
}

func (b *BaseLayout) Render(size *Size) string {
	if b.IsDirty() {
		b.ResizeLayout(size)
		b.SetDirty(false)
	}

	return b.view
}

func (b *BaseLayout) SetDirty(dirty bool) {
	b.dirty = dirty
}

func (b *BaseLayout) SetView(view string) {
	b.view = view
}

func (b *BaseLayout) IsDirty() bool {
	return b.dirty
}

func (b *BaseLayout) GetElements() []Renderable {
	return b.elements
}
