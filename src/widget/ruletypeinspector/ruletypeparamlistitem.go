package ruletypeinspector

import (
	"farental/core/data/api"
	"farental/widget/multivalueselector"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/widget/label"
	"github.com/halsten-dev/orvyn/widget/list"
	"github.com/halsten-dev/orvyn/widget/textinput"
)

type PossibleValueData struct {
	api.ScriptRuleTypePossibleValue
}

func (d PossibleValueData) RenderValue() string {
	return d.Value
}

type ListItem struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	data ParamData

	nameLabel *label.Widget

	inputValue         *textinput.Widget
	multiValueSelector *multivalueselector.Widget[PossibleValueData]

	focusManager *orvyn.FocusManager

	layout *layout.VBoxLayout
}

func Constructor(data ParamData) list.ListItem[ParamData] {
	w := new(ListItem)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.data = data

	w.nameLabel = label.New(w.data.Name)

	w.inputValue = textinput.New()
	w.inputValue.SetActive(false)

	w.multiValueSelector = multivalueselector.New[PossibleValueData]()
	w.multiValueSelector.SetActive(false)

	// The focusManager will be useful to ensure to focus the first active widget when calling focusFirst()
	w.focusManager = orvyn.NewFocusManager()

	w.focusManager.Add(w.inputValue)
	w.focusManager.Add(w.multiValueSelector)

	pileLayout := layout.NewPileLayout(
		[]orvyn.Renderable{
			w.inputValue,
			w.multiValueSelector,
		})

	w.layout = layout.NewMaxWidthVBoxLayout(0,
		[]orvyn.Renderable{
			w.nameLabel,
			pileLayout,
		})

	w.UpdateData(data)

	return w
}

func (w *ListItem) Update(msg tea.Msg) tea.Cmd {
	cmd := w.focusManager.Update(msg)

	w.updateValue()

	return cmd
}

func (w *ListItem) UpdateData(data ParamData) {
	w.data = data

	if len(w.data.PossibleValues) > 0 {
		var keys []string
		var data map[string]PossibleValueData
		var selectedIndex int

		selectedIndex = 0

		data = make(map[string]PossibleValueData)

		w.multiValueSelector.SetActive(true)

		for i, pv := range w.data.PossibleValues {
			keys = append(keys, pv.Key)

			if pv.Key == w.data.Value {
				selectedIndex = i
			}

			data[pv.Key] = PossibleValueData{
				ScriptRuleTypePossibleValue: pv,
			}
		}

		w.multiValueSelector.SetValues(keys, data)

		w.multiValueSelector.SetSelected(selectedIndex)

		return
	}

	w.inputValue.SetActive(true)

	w.inputValue.SetValue(w.data.Value)
}

func (w *ListItem) GetData() ParamData {
	return w.data
}

func (w *ListItem) Resize(size orvyn.Size) {
	size.Height = 4
	w.BaseWidget.Resize(size)
	w.layout.Resize(size)
}

func (w *ListItem) Render() string {
	return w.layout.Render()
}

func (w *ListItem) OnFocus() {
	w.focusManager.FocusFirst()
}

func (w *ListItem) OnBlur() {
	w.focusManager.BlurCurrent()
}

func (w *ListItem) OnEnterInput() {
}

func (w *ListItem) OnExitInput() {
}

func (w *ListItem) FilterValue() string {
	return ""
}

// updateValue updates the Value property of the data based on the active widget.
func (w *ListItem) updateValue() {
	switch {
	case w.multiValueSelector.IsActive():
		w.data.Value = w.multiValueSelector.GetSelectedValue().Key
	case w.inputValue.IsActive():
		w.data.Value = w.inputValue.Value()
	}

}
