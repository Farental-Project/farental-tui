package widget

import (
	"farental/internal/orvyn"
	"farental/style"
	"github.com/charmbracelet/lipgloss"
)

type SimpleListItem struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable
	style lipgloss.Style

	value string
}

func SimpleListItemConstructor(value string) *SimpleListItem {
	sli := new(SimpleListItem)

	sli.value = value

	return sli
}

func (s *SimpleListItem) Render() string {
	return s.value
}

func (s *SimpleListItem) OnFocus() {
	s.style = style.FocusedStyle
}

func (s *SimpleListItem) OnBlur() {
	s.style = style.BlurredStyle
}

func (s *SimpleListItem) OnEnterInput() {}

func (s *SimpleListItem) OnExitInput() {}
