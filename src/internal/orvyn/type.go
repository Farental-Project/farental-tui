package orvyn

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Renderable interface {
	Render() string
	Resize(Size)
	GetSize() Size
	GetMinSize() Size
	GetPreferredSize() Size
	GetMaxSize() Size
	SetVisible(bool)
	IsVisible() bool
}

type BaseRenderable struct {
	size    Size
	visible bool
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

func (b *BaseRenderable) GetMaxSize() Size {
	return NewSize(0, 0)
}

func (b *BaseRenderable) SetVisible(visible bool) {
	b.visible = visible
}

func (b *BaseRenderable) IsVisible() bool {
	return b.visible
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
