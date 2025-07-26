package orvyn

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Renderable interface {
	Render(Size) string
	Resize(Size)
	GetSize() Size
	GetMinSize() Size
	GetPreferredSize() Size
	GetMaxSize() Size
	SetVisible(bool)
	IsVisible() bool
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
