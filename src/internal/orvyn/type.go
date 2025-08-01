package orvyn

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Activable interface {
	SetActive(bool)
	IsActive() bool
}

type BaseActivable struct {
	active bool
}

func (b *BaseActivable) SetActive(active bool) {
	b.active = active
}

func (b *BaseActivable) IsActive() bool {
	return b.active
}

func NewBaseActivable() BaseActivable {
	a := BaseActivable{}

	a.active = true

	return a
}

type Renderable interface {
	Activable

	Render() string
	Resize(Size)
	GetSize() Size
	GetMinSize() Size
	GetPreferredSize() Size
}

type BaseRenderable struct {
	BaseActivable

	size Size
}

func NewBaseRenderable() BaseRenderable {
	b := BaseRenderable{}

	b.BaseActivable = NewBaseActivable()

	return b
}

func (b *BaseRenderable) Resize(size Size) {
	b.size = size
}

func (b *BaseRenderable) GetSize() Size {
	return b.size
}

func (b *BaseRenderable) GetMinSize() Size {
	return NewSize(0, 0)
}

func (b *BaseRenderable) GetPreferredSize() Size {
	return NewSize(0, 0)
}

type Updatable interface {
	Update(tea.Msg) tea.Cmd
}

// Size is a simple struct to represent a size.
type Size struct {
	Width  int
	Height int
}

// NewSize returns a new Size.
func NewSize(width, height int) Size {
	return Size{width, height}
}
