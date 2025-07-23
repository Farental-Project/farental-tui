package errormessage

import (
	"farental/internal/orvyn"
	"farental/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Widget struct {
	errorMsg string
}

func New() *Widget {
	return new(Widget)
}

func (w *Widget) Init() tea.Cmd {
	w.errorMsg = ""

	return nil
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	return nil
}

func (w *Widget) Render(size orvyn.Size) string {
	s := ""

	if w.errorMsg != "" {
		s = style.ErrorStyle.
			Width(size.Width).
			AlignHorizontal(lipgloss.Center).
			Render(w.errorMsg)
	}

	return s
}

func (w *Widget) Resize(size orvyn.Size) {}

func (w *Widget) GetSize() orvyn.Size {
	return orvyn.NewSize(lipgloss.Width(w.errorMsg),
		lipgloss.Height(w.errorMsg))
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

func (w *Widget) SetErrorMsg(msg string) {
	w.errorMsg = msg
}

func (w *Widget) SetError(err error) {
	w.errorMsg = err.Error()
}
