package ruletypeparamlistitem

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

type Data struct {
	api.ScriptRuleTypeParamValue
}

func (d Data) RenderValue() string {
	return d.Value
}

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	data *api.ScriptRuleTypeParam

	nameLabel *label.Widget

	inputValue         *textinput.Widget
	multiValueSelector *multivalueselector.Widget[Data]

	focusManager *orvyn.FocusManager

	layout *layout.VBoxLayout
}

func Constructor(data *api.ScriptRuleTypeParam) list.IListItem {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.data = data

	w.nameLabel = label.New(w.data.Name)

	w.inputValue = textinput.New()
	w.inputValue.SetActive(false)

	w.multiValueSelector = multivalueselector.New[Data]()
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

	w.init()

	return w
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	cmd := w.focusManager.Update(msg)

	return cmd
}

func (w *Widget) Resize(size orvyn.Size) {
	size.Height = 4
	w.BaseWidget.Resize(size)
	w.layout.Resize(size)
}

func (w *Widget) Render() string {
	return w.layout.Render()
}

func (w *Widget) OnFocus() {
	w.focusManager.FocusFirst()
}

func (w *Widget) OnBlur() {
	w.focusManager.BlurCurrent()
}

func (w *Widget) OnEnterInput() {
}

func (w *Widget) OnExitInput() {
}

func (w *Widget) FilterValue() string {
	return ""
}

func (w *Widget) init() {
	if len(w.data.PossibleValues) > 0 {
		var keys []string
		var data map[string]Data

		data = make(map[string]Data)

		w.multiValueSelector.SetActive(true)

		for _, pv := range w.data.PossibleValues {
			keys = append(keys, pv.Key)
			data[pv.Key] = Data{
				ScriptRuleTypeParamValue: pv,
			}
		}

		w.multiValueSelector.SetValues(keys, data)

		return
	}

	w.inputValue.SetActive(true)
}
