package orvyn

import (
	"github.com/charmbracelet/lipgloss"
)

func GetRenderSize(value string) Size {
	width, height := lipgloss.Size(value)

	return NewSize(width, height)
}
