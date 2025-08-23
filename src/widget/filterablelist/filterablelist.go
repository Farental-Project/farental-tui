package filterablelist

import (
	"farental/internal/keybind"
	"farental/style"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
)

type Widget struct {
	orvyn.BaseWidget

	list.Model

	delegate list.ItemDelegate

	MinSize         orvyn.Size
	PreferredSize   orvyn.Size
	CustomEnterDesc string

	NoContentStyle lipgloss.Style
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

	w.NoContentStyle = style.DimTextStyle.Align(lipgloss.Center, lipgloss.Center)

	return w
}

func (w *Widget) Init() tea.Cmd {
	w.Model.ResetSelected()
	w.Model.ResetFilter()

	return nil
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	w.Model, cmd = w.Model.Update(msg)

	w.updateHelp()

	return cmd
}

func (w *Widget) Render() string {
	if len(w.Model.Items()) == 0 {
		size := w.GetSize()
		return w.NoContentStyle.
			Width(size.Width).
			Height(size.Height).
			Render(lokyn.L("Nothing to show"))
	}

	return w.Model.View()

}

func (w *Widget) Resize(size orvyn.Size) {
	w.BaseWidget.Resize(size)

	listItemHeight := w.delegate.Height()
	itemHeight := listItemHeight + style.FocusedStyle.GetVerticalFrameSize()
	itemCount := size.Height / itemHeight

	w.Model.SetWidth(size.Width)
	w.Model.SetHeight(itemCount * listItemHeight)
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
		bubblehelp.ShowAll = false
		bubblehelp.UpdateKeybindHelpDesc(keybind.Esc, lokyn.L("cancel"))
		bubblehelp.UpdateKeybindHelpDesc(keybind.Enter, lokyn.L("apply"))
		bubblehelp.SetKeybindVisible(keybind.Filter, false)
		bubblehelp.SetKeybindVisible(keybind.Help, false)
	case list.FilterApplied:
		bubblehelp.UpdateKeybindHelpDesc(keybind.Esc, lokyn.L("clear filter"))
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
