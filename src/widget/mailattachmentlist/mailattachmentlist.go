package mailattachmentlist

import (
	"farental/internal/keybind"
	"farental/style"
	"farental/widget/list"
	"github.com/charmbracelet/bubbles/key"
	tealist "github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
)

type ShowAttachmentSelectMsg uint

func ShowAttachmentSelectCmd() tea.Msg {
	return ShowAttachmentSelectMsg(1)
}

type DeleteAttachmentMsg int

func DeleteAttachmentCmd(index int) tea.Cmd {
	return func() tea.Msg {
		return DeleteAttachmentMsg(index)
	}
}

type Widget struct {
	orvyn.BaseFocusable

	list.Widget

	contentSize orvyn.Size
}

func New(delegate tealist.ItemDelegate) *Widget {
	keymapContext := bubblehelp.NewKeymap(2)
	keymapContext.Style = style.MainHelpStyle
	keymapContext.NewKeyBinding(keybind.Up, true)
	keymapContext.NewKeyBinding(keybind.Down, true)
	keymapContext.NewKeyBinding(keybind.DKey, true)
	keymapContext.SetHelpDesc(keybind.DKey, lokyn.L("delete"))
	keymapContext.NewKeyBinding(keybind.AKey, true)
	keymapContext.SetHelpDesc(keybind.AKey, lokyn.L("add"))
	keymapContext.NewKeyBinding(keybind.Esc, true)
	keymapContext.SetHelpDesc(keybind.Esc, lokyn.L("stop editing"))

	bubblehelp.RegisterContext(keybind.ContextMailDetailEditorAttachmentList, keymapContext)

	w := new(Widget)

	w.Widget.PreferredSize.Height = 13

	w.Widget = *list.New(delegate, []tealist.Item{})

	return w
}

func (w *Widget) Init() tea.Cmd {
	w.SetItems([]tealist.Item{})

	return nil
}

func (w *Widget) Render() string {
	var s lipgloss.Style

	if w.IsFocused() {
		s = style.FocusedStyle
	} else {
		s = style.BlurredStyle
	}

	return s.Width(w.contentSize.Width).
		Height(w.contentSize.Height).
		Render(w.Widget.Render())
}

func (w *Widget) Resize(size orvyn.Size) {
	size.Width -= style.BlurredStyle.GetHorizontalFrameSize()
	size.Height -= style.BlurredStyle.GetVerticalFrameSize()

	w.contentSize = size
	w.Widget.Resize(size)
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.AKey):
			return ShowAttachmentSelectCmd

		case key.Matches(msg, keybind.DKey):
			return DeleteAttachmentCmd(w.Index())
		}
	}

	cmd := w.Widget.Update(msg)

	return cmd
}

func (w *Widget) OnFocus() {
	bubblehelp.SwitchContext(keybind.ContextMailDetailEditorAttachmentList)
}

func (w *Widget) OnBlur() {
	bubblehelp.SwitchToPreviousContext()
}

func (w *Widget) OnEnterInput() {
}

func (w *Widget) OnExitInput() {
}

func (w *Widget) GetEnterInputKeybind() *key.Binding {
	return &keybind.EKey
}
