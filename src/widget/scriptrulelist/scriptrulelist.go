package scriptrulelist

import (
	"farental/core/data"
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/keybind"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-resty/resty/v2"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/widget/widgetlist"
)

type FocusInspectorMsg int

func FocusInspectorCmd() tea.Msg {
	return FocusInspectorMsg(1)
}

type Widget struct {
	widgetlist.Widget[Data]
	readOnly bool
}

func New() *Widget {
	w := new(Widget)

	w.Widget = *widgetlist.New(Constructor)
	w.Widget.BaseFocusable = orvyn.NewBaseFocusable(w)
	w.Widget.AutoFocusNewItem = true

	return w
}

func (w *Widget) Init() tea.Cmd {
	cmd := w.Widget.Init()

	w.SetItems([]Data{})

	w.FocusFirst()

	w.readOnly = false

	return cmd
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	if m, ok := orvyn.GetKeyMsg(msg); ok && !w.Widget.IsInputting() {
		switch {
		case key.Matches(m, keybind.ShiftUp):
			currentIndex := w.GetGlobalIndex()
			if currentIndex > 0 {
				w.MoveItem(currentIndex, currentIndex-1)
				w.updateRulesOrder()
			}

		case key.Matches(m, keybind.ShiftDown):
			currentIndex := w.GetGlobalIndex()
			if currentIndex < w.Length()-1 {
				w.MoveItem(currentIndex, currentIndex+1)
				w.updateRulesOrder()
			}

		case key.Matches(m, keybind.NKey):
			if bubblehelp.IsKeybindVisible(keybind.NKey) {
				w.Widget.AppendItem(w.getNewRuleData())

				w.updateKeybind()
				return nil
			}

		case key.Matches(m, keybind.IKey):
			if bubblehelp.IsKeybindVisible(keybind.IKey) {
				w.Widget.InsertItem(w.GetGlobalIndex(), w.getNewRuleData())

				w.updateRulesOrder()

				w.updateKeybind()
				return nil
			}

		case key.Matches(m, keybind.CKey):
			if bubblehelp.IsKeybindVisible(keybind.CKey) {
				selectedRule := w.GetSelectedItem()

				selectedRule.Order = 0

				w.Widget.InsertItem(w.GetGlobalIndex(), selectedRule)

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

		case key.Matches(m, keybind.EKeyCtrl):
			return FocusInspectorCmd
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
	bubblehelp.SoftSwitchContext(keybind.ContextScriptEditorRulesList)
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
	var ruleTypeName string
	var ability *api.AbilityResponse
	var resp *resty.Response
	var err error
	var ok bool

	for _, rb := range *data {
		ruleTypeName = ""

		if rb.AbilityCode != "" {
			resp, err = helper.SendRequest(request.AbilityGet(rb.AbilityCode))

			if err != nil {
				return err
			}

			ability, ok = resp.Result().(*api.AbilityResponse)

			if !ok {
				return fmt.Errorf("%s", lokyn.L("Invalid response from server"))
			}
		}

		if rb.RuleTypeCode != "" {
			ruleTypeName, err = w.getRuleTypeName(rb.RuleTypeCode)

			if err != nil {
				return err
			}
		}

		item := Data{
			ScriptRuleBody: rb,
			Ability:        *ability,
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
		bubblehelp.SetKeybindVisible(keybind.CKey, false)
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
	bubblehelp.SetKeybindVisible(keybind.CKey, !limitReached)
}

func (w *Widget) getRuleTypeName(code string) (string, error) {
	resp, err := helper.SendRequest(request.ScriptGetRuleType(code))

	if err != nil {
		return "", err
	}

	ruleType, ok := resp.Result().(*api.ScriptRuleTypeResponse)

	if !ok {
		return "", fmt.Errorf("%s", lokyn.L("Invalid response from server"))
	}

	return ruleType.Name, nil
}

func (w *Widget) getNewRuleData() Data {
	var ruleTypeCode, ruleTypeName string

	ruleTypeCode = "All"
	ruleTypeName, err := w.getRuleTypeName(ruleTypeCode)

	if err != nil {
		ruleTypeCode = ""
		ruleTypeName = ""
	}

	return Data{
		ScriptRuleBody: api.ScriptRuleBody{
			AbilityTarget:  api.TargetSelf,
			RuleTypeTarget: nil,
			Order:          len(w.GetItems()) + 1,
			RuleTypeCode:   ruleTypeCode,
			AbilityCode:    "",
			Parameters:     make([]api.ScriptRuleTypeParam, 0),
		},
		RuleTypeName: ruleTypeName,
	}
}
