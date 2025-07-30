package list

import (
	"farental/internal/orvyn"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type Widget struct {
	orvyn.BaseWidget

	list.Model

	MinSize       orvyn.Size
	PreferredSize orvyn.Size
	MaxSize       orvyn.Size
}

func New(delegate list.ItemDelegate, items []list.Item) *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.Model = list.New(items, delegate, 0, 0)
	w.Model.DisableQuitKeybindings()

	return w
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	w.Model, cmd = w.Model.Update(msg)

	return cmd
}

func (w *Widget) Render() string {
	return w.Model.View()
}

func (w *Widget) Resize(size orvyn.Size) {
	w.BaseWidget.Resize(size)

	w.Model.SetWidth(size.Width)
	w.Model.SetHeight(size.Height)
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
