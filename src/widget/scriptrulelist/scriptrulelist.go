package scriptrulelist

import (
	"farental/core/data"
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

func (w *Widget) Init() tea.Cmd {
	cmd := w.Widget.Init()

	w.SetItems([]api.ScriptRuleBody{})

	return cmd
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	if m, ok := orvyn.GetKeyMsg(msg); ok && !w.Widget.IsInputting() {
		switch {
		case key.Matches(m, keybind.NKey):
			if bubblehelp.IsKeybindVisible(keybind.NKey) {
				w.Widget.AppendItem(api.ScriptRuleBody{
					Target:     api.TargetSelf,
					Order:      len(w.GetItems()) + 1,
					RuleTypeID: 0,
					AbilityID:  0,
					Parameters: "",
				})

				w.updateKeybind()
				return nil
			}

		case key.Matches(m, keybind.IKey):
			if bubblehelp.IsKeybindVisible(keybind.IKey) {
				w.Widget.InsertItem(w.GetGlobalIndex(),
					api.ScriptRuleBody{
						Target:     api.TargetSelf,
						Order:      0,
						RuleTypeID: 0,
						AbilityID:  0,
						Parameters: "",
					})

				w.updateRulesOrder()

				w.updateKeybind()
				return nil
			}

		case key.Matches(m, keybind.DKey):
			w.RemoveItem(w.GetGlobalIndex())

			w.updateRulesOrder()

			w.updateKeybind()

			return nil
		}
	}

	cmd := w.Widget.Update(msg)

	w.updateKeybind()

	return cmd
}

func (w *Widget) updateRulesOrder() {
	items := w.GetItems()

	for i, r := range items {
		r.Order = i + 1
		w.Widget.SetItem(i, r)
	}
}

func (w *Widget) updateKeybind() {
	limitReached := len(w.Widget.GetItems()) == data.ConstScriptMaxRules

	bubblehelp.SetKeybindVisible(keybind.NKey, !limitReached)
	bubblehelp.SetKeybindVisible(keybind.IKey, !limitReached)
}

func (w *Widget) OnFocus() {
	w.Widget.OnFocus()
	bubblehelp.SwitchContext(keybind.ContextScriptEditorRulesList)
	w.updateKeybind()
}

func (w *Widget) OnBlur() {
	w.Widget.OnBlur()
}
