package orvyn

import (
	"github.com/charmbracelet/lipgloss"
)

type StaticRenderable struct {
	view string
}

func NewStaticRenderable(staticContent string) *StaticRenderable {
	return &StaticRenderable{view: staticContent}
}

func (s *StaticRenderable) Render(size Size) string {
	return s.view
}

func (s *StaticRenderable) Resize(size Size) {}

func (s *StaticRenderable) GetSize() Size {
	return NewSize(lipgloss.Width(s.view), lipgloss.Height(s.view))
}

func (s *StaticRenderable) GetMinSize() Size {
	return NewSize(0, 0)
}

func (s *StaticRenderable) GetPreferredSize() Size {
	return NewSize(0, 0)
}

func (s *StaticRenderable) GetMaxSize() Size {
	return NewSize(0, 0)
}
