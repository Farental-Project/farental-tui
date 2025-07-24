package orvyn

import (
	"github.com/charmbracelet/lipgloss"
)

type SimpleRenderable struct {
	view string
}

var VGap = NewSimpleRenderable("\n")

func NewSimpleRenderable(view string) *SimpleRenderable {
	return &SimpleRenderable{view: view}
}

func (s *SimpleRenderable) SetView(view string) {
	s.view = view
}

func (s *SimpleRenderable) Render(size Size) string {
	return s.view
}

func (s *SimpleRenderable) Resize(size Size) {}

func (s *SimpleRenderable) GetSize() Size {
	return NewSize(lipgloss.Width(s.view), lipgloss.Height(s.view))
}

func (s *SimpleRenderable) GetMinSize() Size {
	return NewSize(0, 0)
}

func (s *SimpleRenderable) GetPreferredSize() Size {
	return NewSize(0, 0)
}

func (s *SimpleRenderable) GetMaxSize() Size {
	return NewSize(0, 0)
}
