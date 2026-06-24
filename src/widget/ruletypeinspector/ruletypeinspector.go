package ruletypeinspector

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/keybind"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/widgetlist"
)

type FocusRuleListMsg int

func FocusRuleListCmd() tea.Msg {
	return FocusRuleListMsg(1)
}

type ParamData struct {
	api.ScriptRuleTypeStructParam
	Value string
}

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	parameters *widgetlist.Widget[ParamData]

	noParamText *orvyn.SimpleRenderable
}

func New() *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()
	w.BaseFocusable = orvyn.NewBaseFocusable(w)

	w.parameters = widgetlist.New(Constructor)
	w.parameters.SetFilterable(false)
	w.parameters.SetCursorMovementKeybinds(keybind.ShiftTab, keybind.Tab)
	w.parameters.InfiniteScroll = true
	w.parameters.CursorMovedCallback = w.cursorMoved

	w.noParamText = orvyn.NewSimpleRenderable("")
	w.noParamText.SizeConstraint = true
	w.noParamText.Style = orvyn.GetTheme().Style(theme.DimTextStyleID)

	w.OnBlur()

	return w
}

func (w *Widget) Init() tea.Cmd {
	w.SetRuleType("", nil)

	cmd := w.parameters.Init()

	w.noParamText.SetValue(lokyn.L("No parameters"))

	w.SetActive(!w.IsEmpty())

	return cmd
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	if k, ok := orvyn.GetKeyMsg(msg); ok {
		switch {
		case key.Matches(k, keybind.EKeyCtrl):
			return FocusRuleListCmd
		}
	}

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
	ws := orvyn.GetTheme().Style(theme.BlurredWidgetStyleID)

	if w.IsEmpty() {
		contentSize := w.GetContentSize()
		return ws.Width(contentSize.Width).
			Height(contentSize.Height).
			Render(w.noParamText.Render())
	} else {
		return w.parameters.Render()
	}
}

func (w *Widget) OnFocus() {
	w.parameters.OnFocus()
	bubblehelp.SoftSwitchContext(keybind.ContextScriptEditorRuleInspectorNormalMode)
}

func (w *Widget) OnBlur() {
	w.parameters.OnBlur()
}

func (w *Widget) OnEnterInput() tea.Cmd {
	w.parameters.FocusFirst()
	bubblehelp.SwitchContext(keybind.ContextScriptEditorRuleInspector)

	return nil
}

func (w *Widget) OnExitInput() tea.Cmd {
	w.parameters.BlurCurrent()
	bubblehelp.SwitchToPreviousContext()

	return nil
}

func (w *Widget) GetFocusKeybind() *key.Binding {
	return &keybind.Num3Key
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

	ruleParamStruct, err := helper.Fetch[[]api.ScriptRuleTypeStructParam](request.ScriptGetRuleTypeParamStruct(code))

	if err != nil {
		return err
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

	w.SetActive(!w.IsEmpty())

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

func (w *Widget) IsEmpty() bool {
	if w.parameters.Length() == 0 {
		return true
	}

	return false
}
