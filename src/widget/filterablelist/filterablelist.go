package filterablelist

import (
	"farental/internal/keybind"
	"farental/internal/lang"
	"farental/internal/orvyn"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
)

type Widget struct {
	orvyn.BaseWidget

	list.Model

	delegate list.ItemDelegate

	MinSize         orvyn.Size
	PreferredSize   orvyn.Size
	CustomEnterDesc string
}

func New(delegate list.ItemDelegate, items []list.Item) *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.delegate = delegate

	w.Model = list.New(items, delegate, 0, 0)
	w.Model.DisableQuitKeybindings()

	w.Model.SetShowStatusBar(false)
	w.Model.SetShowHelp(false)
	w.Model.SetShowTitle(false)
	w.Model.SetShowPagination(true)

	return w
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	w.Model, cmd = w.Model.Update(msg)

	w.updateHelp()

	return cmd
}

func (w *Widget) Render() string {
	return w.Model.View()
}

func (w *Widget) Resize(size orvyn.Size) {
	w.BaseWidget.Resize(size)

	itemHeight := w.delegate.Height()
	itemCount := size.Height / itemHeight
	itemCount -= 2

	w.Model.SetWidth(size.Width)
	w.Model.SetHeight(itemCount * itemHeight)
}

func (w *Widget) GetMinSize() orvyn.Size {
	return w.MinSize
}

func (w *Widget) GetPreferredSize() orvyn.Size {
	return w.PreferredSize
}

func (w *Widget) updateHelp() {
	switch w.FilterState() {
	case list.Filtering:
		bubblehelp.UpdateKeybindHelpDesc(keybind.Esc, lang.L("cancel"))
		bubblehelp.UpdateKeybindHelpDesc(keybind.Enter, lang.L("apply"))
		bubblehelp.SetKeybindVisible(keybind.Filter, false)
		bubblehelp.SetKeybindVisible(keybind.Help, false)
	case list.FilterApplied:
		bubblehelp.UpdateKeybindHelpDesc(keybind.Esc, lang.L("clear filter"))
		bubblehelp.UpdateKeybindHelpDesc(keybind.Enter, w.CustomEnterDesc)
		bubblehelp.SetKeybindVisible(keybind.Filter, true)
		bubblehelp.SetKeybindVisible(keybind.Help, true)
	case list.Unfiltered:
		bubblehelp.UpdateKeybindHelpDesc(keybind.Esc, "")
		bubblehelp.UpdateKeybindHelpDesc(keybind.Enter, w.CustomEnterDesc)
		bubblehelp.SetKeybindVisible(keybind.Filter, true)
		bubblehelp.SetKeybindVisible(keybind.Help, true)
	}

}
