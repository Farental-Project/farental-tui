package label

import (
	"farental/internal/orvyn"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Widget struct {
	Style lipgloss.Style

	value string
}

func New(value string) *Widget {
	w := new(Widget)

	w.Style = lipgloss.NewStyle().Italic(true)
	w.value = value

	return w
}

func (w *Widget) SetValue(value string) {
	w.value = value
}

func (w *Widget) Init() tea.Cmd {
	return nil
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	return nil
}

func (w *Widget) Render(size orvyn.Size) string {
	return w.Style.
		Width(size.Width).
		Height(size.Height).
		Render(w.value)
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
