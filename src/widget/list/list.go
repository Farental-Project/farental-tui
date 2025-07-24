package list

import (
	"farental/internal/orvyn"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type Widget struct {
	list.Model

	MinSize       orvyn.Size
	PreferredSize orvyn.Size
	MaxSize       orvyn.Size
}

func New(delegate list.ItemDelegate, items []list.Item) *Widget {
	w := new(Widget)

	w.Model = list.New(items, delegate, 0, 0)
	w.Model.DisableQuitKeybindings()

	return w
}

func (w *Widget) Init() tea.Cmd {
	return nil
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	w.Model, cmd = w.Model.Update(msg)

	return cmd
}

func (w *Widget) Render(size orvyn.Size) string {
	return w.Model.View()
}

func (w *Widget) Resize(size orvyn.Size) {
	w.Model.SetWidth(size.Width)
	w.Model.SetHeight(size.Height)
}

func (w *Widget) GetSize() orvyn.Size {
	return orvyn.NewSize(w.Model.Width(), w.Model.Height())
}

func (w *Widget) GetMinSize() orvyn.Size {
	return w.MinSize
}

func (w *Widget) GetPreferredSize() orvyn.Size {
	return w.PreferredSize
}

func (w *Widget) GetMaxSize() orvyn.Size {
	return w.MaxSize
}
