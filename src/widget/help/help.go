package help

import (
	"farental/internal/orvyn"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
)

type Widget struct{}

func New() *Widget {
	return new(Widget)
}

func (w *Widget) Init() tea.Cmd {
	return nil
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	return nil
}

func (w *Widget) Render(size orvyn.Size) string {
	return bubblehelp.View(size.Width)
}

func (w *Widget) Resize(size orvyn.Size) {}

func (w *Widget) GetSize() orvyn.Size {
	return orvyn.NewSize(0, 0)
}

func (w *Widget) GetMinSize() orvyn.Size {
	return orvyn.NewSize(0, 0)
}

func (w *Widget) GetPreferredSize() orvyn.Size {
	return orvyn.NewSize(0, 0)
}

func (w *Widget) GetMaxSize() orvyn.Size {
	return orvyn.NewSize(0, 0)
}
