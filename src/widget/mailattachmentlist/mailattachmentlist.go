package mailattachmentlist

import (
	"farental/core/data/api"
	"farental/internal/keybind"
	"farental/internal/style"
	"farental/widget/mailattachmentlistitem"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/widget/list"
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
	list.Widget[api.StackResponse]
}

func New() *Widget {
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

	w.Widget = *list.New(mailattachmentlistitem.Constructor)

	w.OnBlur()

	return w
}

func (w *Widget) Init() tea.Cmd {
	w.SetItems([]api.StackResponse{})

	return nil
}

func (w *Widget) Render() string {
	return w.Render()
}

func (w *Widget) Resize(size orvyn.Size) {
	w.Widget.Resize(size)
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.AKey):
			return ShowAttachmentSelectCmd

		case key.Matches(msg, keybind.DKey):
			return DeleteAttachmentCmd(w.GetGlobalIndex())
		}
	}

	cmd := w.Widget.Update(msg)

	return cmd
}

func (w *Widget) OnFocus() {
	w.Widget.OnFocus()
	bubblehelp.SwitchContext(keybind.ContextMailDetailEditorAttachmentList)
}

func (w *Widget) OnBlur() {
	w.Widget.OnBlur()
	bubblehelp.SwitchToPreviousContext()
}

func (w *Widget) OnEnterInput() {
}

func (w *Widget) OnExitInput() {
}

func (w *Widget) GetEnterInputKeybind() *key.Binding {
	return &keybind.EKey
}
