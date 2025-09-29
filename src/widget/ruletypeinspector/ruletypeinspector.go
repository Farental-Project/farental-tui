package ruletypeinspector

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/widget/ruletypeparamlistitem"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/widget/list"
)

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	parameters *list.Widget[api.ScriptRuleTypeParam]

	contentSize orvyn.Size
}

func New() *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.parameters = list.New(ruletypeparamlistitem.Constructor)
	w.parameters.SetFilterable(false)
	w.parameters.CursorMovedCallback = w.cursorMoved

	w.OnBlur()

	return w
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
func (w *Widget) SetRuleType(code string) error {
	if code == "" {
		w.parameters.SetItems([]api.ScriptRuleTypeParam{})
		w.parameters.SetActive(false)
		return nil
	}

	resp, err := helper.SendRequest(request.ScriptGetRuleTypeParamStruct(code))

	if err != nil {
		return err
	}

	ruleParamStruct, ok := resp.Result().(*api.ScriptRuleTypeParamStructResponse)

	if !ok {
		return fmt.Errorf(lokyn.L("Invalid response"))
	}

	w.parameters.SetItems(ruleParamStruct.Parameters)
	w.parameters.BlurCurrent()

	return nil
}

// UpdateData reads the script response and set parameters values.
func (w *Widget) UpdateData(data *api.ScriptResponse) {

}

// GetData returns the parameters as a json string.
func (w *Widget) GetData() string {
	// TODO: export parameters as json string
	return ""
}
