package orvyn

import (
	"github.com/charmbracelet/lipgloss"
)

type SimpleRenderable struct {
	BaseRenderable

	Style lipgloss.Style

	value string
}

var VGap = NewSimpleRenderable("\n")

func NewSimpleRenderable(value string) *SimpleRenderable {
	s := new(SimpleRenderable)

	s.BaseRenderable = NewBaseRenderable()
	s.value = value
	s.Style = lipgloss.NewStyle()

	return s
}

func (s *SimpleRenderable) SetValue(value string) {
	s.value = value
}

func (s *SimpleRenderable) Render() string {
	return s.Style.Render(s.value)
}

func (s *SimpleRenderable) GetMinSize() Size {
	return GetRenderSize(s.Style, s.value)
}

func (s *SimpleRenderable) GetPreferredSize() Size {
	return GetRenderSize(s.Style, s.value)
}
