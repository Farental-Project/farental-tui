package scriptrulelist

import (
	"farental/core/data"
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/internal/style"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-resty/resty/v2"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/widget/list"
)

type Widget struct {
	list.Widget[Data]
	readOnly bool
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

	w.Widget = *list.New(Constructor)
	w.Widget.BaseFocusable = orvyn.NewBaseFocusable(w)

	return w
}

func (w *Widget) Init() tea.Cmd {
	cmd := w.Widget.Init()

	w.SetItems([]Data{})

	w.readOnly = false

	return cmd
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	if m, ok := orvyn.GetKeyMsg(msg); ok && !w.Widget.IsInputting() {
		switch {
		case key.Matches(m, keybind.NKey):
			if bubblehelp.IsKeybindVisible(keybind.NKey) {
				w.Widget.AppendItem(
					Data{
						ScriptRuleBody: api.ScriptRuleBody{
							Target:       api.TargetSelf,
							Order:        len(w.GetItems()) + 1,
							RuleTypeCode: "",
							AbilityCode:  "",
							Parameters:   make([]api.ScriptRuleTypeParam, 0),
						},
					})

				w.updateKeybind()
				return nil
			}

		case key.Matches(m, keybind.IKey):
			if bubblehelp.IsKeybindVisible(keybind.IKey) {
				w.Widget.InsertItem(w.GetGlobalIndex(),
					Data{
						ScriptRuleBody: api.ScriptRuleBody{
							Target:       api.TargetSelf,
							Order:        0,
							RuleTypeCode: "",
							AbilityCode:  "",
							Parameters:   make([]api.ScriptRuleTypeParam, 0),
						},
					})

				w.updateRulesOrder()

				w.updateKeybind()
				return nil
			}

		case key.Matches(m, keybind.DKey):
			if bubblehelp.IsKeybindVisible(keybind.DKey) {
				w.RemoveItem(w.GetGlobalIndex())

				w.updateRulesOrder()

				w.updateKeybind()

				return nil
			}

		case key.Matches(m, keybind.EKey):
			if !bubblehelp.IsKeybindVisible(keybind.EKey) {
				return nil
			}
		}
	}

	cmd := w.Widget.Update(msg)

	w.updateKeybind()

	return cmd
}

func (w *Widget) GetFocusKeybind() *key.Binding {
	return &keybind.Num2Key
}

func (w *Widget) OnFocus() {
	w.Widget.OnFocus()
	bubblehelp.SwitchContext(keybind.ContextScriptEditorRulesList)
	w.updateKeybind()
}

func (w *Widget) OnBlur() {
	w.Widget.OnBlur()
}

func (w *Widget) updateRulesOrder() {
	items := w.GetItems()

	for i, r := range items {
		r.Order = i + 1
		w.Widget.SetItem(i, r)
	}
}

func (w *Widget) SetData(data *[]api.ScriptRuleBody) error {
	var listItems []Data
	var abilityName, ruleTypeName string
	var resp *resty.Response
	var err error

	for _, rb := range *data {
		abilityName = ""
		ruleTypeName = ""

		if rb.AbilityCode != "" {
			resp, err = helper.SendRequest(request.AbilityGet(rb.AbilityCode))

			if err != nil {
				return err
			}

			ability, ok := resp.Result().(*api.AbilityResponse)

			if !ok {
				return fmt.Errorf(lokyn.L("Invalid response from server"))
			}

			abilityName = ability.Name
		}

		if rb.RuleTypeCode != "" {
			resp, err = helper.SendRequest(request.ScriptGetRuleType(rb.RuleTypeCode))

			if err != nil {
				return err
			}

			ruleType, ok := resp.Result().(*api.ScriptRuleTypeResponse)

			if !ok {
				return fmt.Errorf(lokyn.L("Invalid response from server"))
			}

			ruleTypeName = ruleType.Name
		}

		item := Data{
			ScriptRuleBody: rb,
			AbilityName:    abilityName,
			RuleTypeName:   ruleTypeName,
		}

		listItems = append(listItems, item)
	}

	w.SetItems(listItems)

	return nil
}

func (w *Widget) SetReadOnly() {
	w.readOnly = true
}

func (w *Widget) updateKeybind() {
	if w.readOnly {
		bubblehelp.SetKeybindVisible(keybind.NKey, false)
		bubblehelp.SetKeybindVisible(keybind.IKey, false)
		bubblehelp.SetKeybindVisible(keybind.EKey, false)
		bubblehelp.SetKeybindVisible(keybind.DKey, false)
		bubblehelp.SetKeybindVisible(keybind.Tab, false)
		bubblehelp.SetKeybindVisible(keybind.ShiftTab, false)
		bubblehelp.SetKeybindVisible(keybind.SKeyCtrl, false)
		bubblehelp.SetKeybindVisible(keybind.Help, false)
		return
	}

	limitReached := len(w.Widget.GetItems()) == data.ConstScriptMaxRules

	bubblehelp.SetKeybindVisible(keybind.NKey, !limitReached)
	bubblehelp.SetKeybindVisible(keybind.IKey, !limitReached)
}
