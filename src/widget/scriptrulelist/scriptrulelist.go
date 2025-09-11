package scriptrulelist

import (
	"farental/core/data/api"
	"farental/internal/keybind"
	"farental/internal/style"
	"farental/widget/scriptrulelistitem"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/widget/list"
)

type Widget struct {
	list.Widget[api.ScriptRuleBody]

	data []api.ScriptRuleBody
}

func New() *Widget {
	listKeymap := bubblehelp.NewKeymap(2)
	listKeymap.Style = style.MainHelpStyle
	listKeymap.NewKeyBinding(keybind.Tab, true)
	listKeymap.NewKeyBinding(keybind.ShiftTab, true)
	listKeymap.NewKeyBinding(keybind.Esc, true)
	listKeymap.SetHelpDesc(keybind.Esc, lokyn.L("stop editing"))
	listKeymap.NewKeyBinding(keybind.Quit, false)

	bubblehelp.RegisterContext(keybind.ContextBasicEditMode, listKeymap)

	w := new(Widget)

	w.Widget = *list.New(scriptrulelistitem.Constructor)

	return w
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	if m, ok := orvyn.GetKeyMsg(msg); ok {
		switch {
		case key.Matches(m, keybind.NKey):

			return nil

		case key.Matches(m, keybind.IKey):

			return nil

		}
	}

	return nil
}

func (w *Widget) OnFocus() {
	w.Widget.OnFocus()
	bubblehelp.SwitchContext(keybind.ContextScriptEditorRulesList)
}

func (w *Widget) OnBlur() {
	w.Widget.OnBlur()
}
