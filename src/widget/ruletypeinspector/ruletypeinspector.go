package ruletypeinspector

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/keybind"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/widget/list"
)

type ParamData struct {
	api.ScriptRuleTypeStructParam
	Value string
}

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	parameters *list.Widget[ParamData]

	contentSize orvyn.Size
}

func New() *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.parameters = list.New(Constructor)
	w.parameters.SetFilterable(false)
	w.parameters.CursorMovedCallback = w.cursorMoved

	w.OnBlur()

	return w
}

func (w *Widget) Init() tea.Cmd {
	w.SetRuleType("", nil)

	cmd := w.parameters.Init()

	return cmd
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	if w.IsInputting() {
		cmd = w.parameters.Update(msg)
	}

	return cmd
}

func (w *Widget) Resize(size orvyn.Size) {
	w.BaseWidget.Resize(size)
	w.parameters.Resize(size)
}

func (w *Widget) Render() string {
	return w.parameters.Render()
}

func (w *Widget) OnFocus() {
	w.parameters.OnFocus()
}

func (w *Widget) OnBlur() {
	w.parameters.OnBlur()
}

func (w *Widget) OnEnterInput() {
	w.parameters.FocusFirst()
}

func (w *Widget) OnExitInput() {
	w.parameters.BlurCurrent()
}

func (w *Widget) GetEnterInputKeybind() *key.Binding {
	return &keybind.EKey
}

func (w *Widget) cursorMoved(index int) {

}

// SetRuleType initialize the inspector based on the given rule type code.
func (w *Widget) SetRuleType(code string, data *[]api.ScriptRuleTypeParam) error {
	var paramData []ParamData

	if code == "" {
		w.parameters.SetItems([]ParamData{})
		w.parameters.SetActive(false)
		return nil
	}

	resp, err := helper.SendRequest(request.ScriptGetRuleTypeParamStruct(code))

	if err != nil {
		return err
	}

	ruleParamStruct, ok := resp.Result().(*[]api.ScriptRuleTypeStructParam)

	if !ok {
		return fmt.Errorf(lokyn.L("Invalid response"))
	}

	paramData = make([]ParamData, 0)

	for _, d := range *ruleParamStruct {
		value := ""

		if data != nil {
			for _, v := range *data {
				if v.Identifier == d.Identifier {
					value = v.Value
					break
				}
			}
		}

		param := ParamData{
			ScriptRuleTypeStructParam: d,
			Value:                     value,
		}

		paramData = append(paramData, param)
	}

	w.parameters.SetItems(paramData)
	w.parameters.BlurCurrent()

	return nil
}

// GetItemsData returns the parameters
func (w *Widget) GetItemsData() []api.ScriptRuleTypeParam {
	var items []api.ScriptRuleTypeParam

	for _, rtp := range w.parameters.GetItems() {
		item := api.ScriptRuleTypeParam{
			Identifier: rtp.Identifier,
			Value:      rtp.Value,
		}

		items = append(items, item)
	}

	return items
}
