package mailattachmentlist

import (
	"farental/internal/keybind"
	"farental/internal/lang"
	"farental/internal/orvyn"
	"farental/style"
	"farental/widget/list"
	"github.com/charmbracelet/bubbles/key"
	tealist "github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/bubblehelp"
)

type ShowAttachmentSelectMsg uint

func ShowAttachmentSelectCmd() tea.Msg {
	return ShowAttachmentSelectMsg(1)
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
	keymapContext.SetHelpDesc(keybind.DKey, lang.L("delete"))
	keymapContext.NewKeyBinding(keybind.AKey, true)
	keymapContext.SetHelpDesc(keybind.AKey, lang.L("add"))
	keymapContext.NewKeyBinding(keybind.Esc, true)
	keymapContext.SetHelpDesc(keybind.Esc, lang.L("stop editing"))

	bubblehelp.RegisterContext(keybind.ContextMailDetailEditorAttachmentList, keymapContext)

	w := new(Widget)

	w.Widget.PreferredSize.Height = 13

	w.Widget = *list.New(delegate, []tealist.Item{})

	return w
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
	w.Widget.Resize(size)

	size.Width -= style.BlurredStyle.GetHorizontalFrameSize()
	size.Height -= style.BlurredStyle.GetVerticalFrameSize()

	w.contentSize = size
	w.SetWidth(size.Width)
	w.SetHeight(size.Height)
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.AKey):
			return ShowAttachmentSelectCmd
		}
	}

	return nil
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
