package ruleinspector

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn"
)

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	style lipgloss.Style
}
